// +build byollvm

// This file has been copied from the LLVM project and was modified for
// use in TinyGo. The below comment is the original header:

//===-- cc1as_main.cpp - Clang Assembler  ---------------------------------===//
//
//                     The LLVM Compiler Infrastructure
//
// This file is distributed under the University of Illinois Open Source
// License. See LICENSE.TXT for details.
//
//===----------------------------------------------------------------------===//
//
// This is the entry point to the clang -cc1as functionality, which implements
// the direct interface to the LLVM MC based assembler.
//
//===----------------------------------------------------------------------===//

#include "clang/Basic/Diagnostic.h"
#include "clang/Basic/DiagnosticOptions.h"
#include "clang/Driver/DriverDiagnostic.h"
#include "clang/Driver/Options.h"
#include "clang/Frontend/FrontendDiagnostic.h"
#include "clang/Frontend/TextDiagnosticPrinter.h"
#include "clang/Frontend/Utils.h"
#include "llvm/ADT/STLExtras.h"
#include "llvm/ADT/StringSwitch.h"
#include "llvm/ADT/Triple.h"
#include "llvm/IR/DataLayout.h"
#include "llvm/MC/MCAsmBackend.h"
#include "llvm/MC/MCAsmInfo.h"
#include "llvm/MC/MCCodeEmitter.h"
#include "llvm/MC/MCContext.h"
#include "llvm/MC/MCInstrInfo.h"
#include "llvm/MC/MCObjectFileInfo.h"
#include "llvm/MC/MCObjectWriter.h"
#include "llvm/MC/MCParser/MCAsmParser.h"
#include "llvm/MC/MCParser/MCTargetAsmParser.h"
#include "llvm/MC/MCRegisterInfo.h"
#include "llvm/MC/MCStreamer.h"
#include "llvm/MC/MCSubtargetInfo.h"
#include "llvm/MC/MCTargetOptions.h"
#include "llvm/Option/Arg.h"
#include "llvm/Option/ArgList.h"
#include "llvm/Option/OptTable.h"
#include "llvm/Support/CommandLine.h"
#include "llvm/Support/ErrorHandling.h"
#include "llvm/Support/Host.h"
#include "llvm/Support/MemoryBuffer.h"
#include "llvm/Support/Path.h"
#include "llvm/Support/Signals.h"
#include "llvm/Support/SourceMgr.h"
#include "llvm/Support/TargetRegistry.h"
#include <memory>
#include <system_error>
using namespace clang;
using namespace clang::driver;
using namespace clang::driver::options;
using namespace llvm;
using namespace llvm::opt;

#include "cc1as.h"

bool AssemblerInvocation::CreateFromArgs(AssemblerInvocation &Opts,
                                         ArrayRef<const char *> Argv,
                                         DiagnosticsEngine &Diags) {
  bool Success = true;

  // Parse the arguments.
  std::unique_ptr<OptTable> OptTbl(createDriverOptTable());

  const unsigned IncludedFlagsBitmask = options::CC1AsOption;
  unsigned MissingArgIndex, MissingArgCount;
  InputArgList Args = OptTbl->ParseArgs(Argv, MissingArgIndex, MissingArgCount,
                                        IncludedFlagsBitmask);

  // Check for missing argument error.
  if (MissingArgCount) {
    Diags.Report(diag::err_drv_missing_argument)
        << Args.getArgString(MissingArgIndex) << MissingArgCount;
    Success = false;
  }

  // Issue errors on unknown arguments.
  for (const Arg *A : Args.filtered(OPT_UNKNOWN)) {
    auto ArgString = A->getAsString(Args);
    std::string Nearest;
    if (OptTbl->findNearest(ArgString, Nearest, IncludedFlagsBitmask) > 1)
      Diags.Report(diag::err_drv_unknown_argument) << ArgString;
    else
      Diags.Report(diag::err_drv_unknown_argument_with_suggestion)
          << ArgString << Nearest;
    Success = false;
  }

  // Construct the invocation.

  // Target Options
  Opts.Triple = llvm::Triple::normalize(Args.getLastArgValue(OPT_triple));
  Opts.CPU = Args.getLastArgValue(OPT_target_cpu);
  Opts.Features = Args.getAllArgValues(OPT_target_feature);

  // Use the default target triple if unspecified.
  if (Opts.Triple.empty())
    Opts.Triple = llvm::sys::getDefaultTargetTriple();

  // Language Options
  Opts.IncludePaths = Args.getAllArgValues(OPT_I);
  Opts.NoInitialTextSection = Args.hasArg(OPT_n);
  Opts.SaveTemporaryLabels = Args.hasArg(OPT_msave_temp_labels);
  // Any DebugInfoKind implies GenDwarfForAssembly.
  Opts.GenDwarfForAssembly = Args.hasArg(OPT_debug_info_kind_EQ);

  if (const Arg *A = Args.getLastArg(OPT_compress_debug_sections,
                                     OPT_compress_debug_sections_EQ)) {
    if (A->getOption().getID() == OPT_compress_debug_sections) {
      // TODO: be more clever about the compression type auto-detection
      Opts.CompressDebugSections = llvm::DebugCompressionType::GNU;
    } else {
      Opts.CompressDebugSections =
          llvm::StringSwitch<llvm::DebugCompressionType>(A->getValue())
              .Case("none", llvm::DebugCompressionType::None)
              .Case("zlib", llvm::DebugCompressionType::Z)
              .Case("zlib-gnu", llvm::DebugCompressionType::GNU)
              .Default(llvm::DebugCompressionType::None);
    }
  }

  Opts.RelaxELFRelocations = Args.hasArg(OPT_mrelax_relocations);
  Opts.DwarfVersion = getLastArgIntValue(Args, OPT_dwarf_version_EQ, 2, Diags);
  Opts.DwarfDebugFlags = Args.getLastArgValue(OPT_dwarf_debug_flags);
  Opts.DwarfDebugProducer = Args.getLastArgValue(OPT_dwarf_debug_producer);
  Opts.DebugCompilationDir = Args.getLastArgValue(OPT_fdebug_compilation_dir);
  Opts.MainFileName = Args.getLastArgValue(OPT_main_file_name);

  for (const auto &Arg : Args.getAllArgValues(OPT_fdebug_prefix_map_EQ))
    Opts.DebugPrefixMap.insert(StringRef(Arg).split('='));

  // Frontend Options
  if (Args.hasArg(OPT_INPUT)) {
    bool First = true;
    for (const Arg *A : Args.filtered(OPT_INPUT)) {
      if (First) {
        Opts.InputFile = A->getValue();
        First = false;
      } else {
        Diags.Report(diag::err_drv_unknown_argument) << A->getAsString(Args);
        Success = false;
      }
    }
  }
  Opts.LLVMArgs = Args.getAllArgValues(OPT_mllvm);
  Opts.OutputPath = Args.getLastArgValue(OPT_o);
  Opts.SplitDwarfFile = Args.getLastArgValue(OPT_split_dwarf_file);
  if (Arg *A = Args.getLastArg(OPT_filetype)) {
    StringRef Name = A->getValue();
    unsigned OutputType = StringSwitch<unsigned>(Name)
      .Case("asm", FT_Asm)
      .Case("null", FT_Null)
      .Case("obj", FT_Obj)
      .Default(~0U);
    if (OutputType == ~0U) {
      Diags.Report(diag::err_drv_invalid_value) << A->getAsString(Args) << Name;
      Success = false;
    } else
      Opts.OutputType = FileType(OutputType);
  }
  Opts.ShowHelp = Args.hasArg(OPT_help);
  Opts.ShowVersion = Args.hasArg(OPT_version);

  // Transliterate Options
  Opts.OutputAsmVariant =
      getLastArgIntValue(Args, OPT_output_asm_variant, 0, Diags);
  Opts.ShowEncoding = Args.hasArg(OPT_show_encoding);
  Opts.ShowInst = Args.hasArg(OPT_show_inst);

  // Assemble Options
  Opts.RelaxAll = Args.hasArg(OPT_mrelax_all);
  Opts.NoExecStack = Args.hasArg(OPT_mno_exec_stack);
  Opts.FatalWarnings = Args.hasArg(OPT_massembler_fatal_warnings);
  Opts.RelocationModel = Args.getLastArgValue(OPT_mrelocation_model, "pic");
  Opts.IncrementalLinkerCompatible =
      Args.hasArg(OPT_mincremental_linker_compatible);
  Opts.SymbolDefs = Args.getAllArgValues(OPT_defsym);

  return Success;
}

static std::unique_ptr<raw_fd_ostream>
getOutputStream(StringRef Path, DiagnosticsEngine &Diags, bool Binary) {
  // Make sure that the Out file gets unlinked from the disk if we get a
  // SIGINT.
  if (Path != "-")
    sys::RemoveFileOnSignal(Path);

  std::error_code EC;
  auto Out = llvm::make_unique<raw_fd_ostream>(
      Path, EC, (Binary ? sys::fs::F_None : sys::fs::F_Text));
  if (EC) {
    Diags.Report(diag::err_fe_unable_to_open_output) << Path << EC.message();
    return nullptr;
  }

  return Out;
}

bool ExecuteAssembler(AssemblerInvocation &Opts, DiagnosticsEngine &Diags) {
  // Get the target specific parser.
  std::string Error;
  const Target *TheTarget = TargetRegistry::lookupTarget(Opts.Triple, Error);
  if (!TheTarget)
    return Diags.Report(diag::err_target_unknown_triple) << Opts.Triple;

  ErrorOr<std::unique_ptr<MemoryBuffer>> Buffer =
      MemoryBuffer::getFileOrSTDIN(Opts.InputFile);

  if (std::error_code EC = Buffer.getError()) {
    Error = EC.message();
    return Diags.Report(diag::err_fe_error_reading) << Opts.InputFile;
  }

  SourceMgr SrcMgr;

  // Tell SrcMgr about this buffer, which is what the parser will pick up.
  SrcMgr.AddNewSourceBuffer(std::move(*Buffer), SMLoc());

  // Record the location of the include directories so that the lexer can find
  // it later.
  SrcMgr.setIncludeDirs(Opts.IncludePaths);

  std::unique_ptr<MCRegisterInfo> MRI(TheTarget->createMCRegInfo(Opts.Triple));
  assert(MRI && "Unable to create target register info!");

  std::unique_ptr<MCAsmInfo> MAI(TheTarget->createMCAsmInfo(*MRI, Opts.Triple));
  assert(MAI && "Unable to create target asm info!");

  // Ensure MCAsmInfo initialization occurs before any use, otherwise sections
  // may be created with a combination of default and explicit settings.
  MAI->setCompressDebugSections(Opts.CompressDebugSections);

  MAI->setRelaxELFRelocations(Opts.RelaxELFRelocations);

  bool IsBinary = Opts.OutputType == AssemblerInvocation::FT_Obj;
  if (Opts.OutputPath.empty())
    Opts.OutputPath = "-";
  std::unique_ptr<raw_fd_ostream> FDOS =
      getOutputStream(Opts.OutputPath, Diags, IsBinary);
  if (!FDOS)
    return true;
  std::unique_ptr<raw_fd_ostream> DwoOS;
  if (!Opts.SplitDwarfFile.empty())
    DwoOS = getOutputStream(Opts.SplitDwarfFile, Diags, IsBinary);

  // FIXME: This is not pretty. MCContext has a ptr to MCObjectFileInfo and
  // MCObjectFileInfo needs a MCContext reference in order to initialize itself.
  std::unique_ptr<MCObjectFileInfo> MOFI(new MCObjectFileInfo());

  MCContext Ctx(MAI.get(), MRI.get(), MOFI.get(), &SrcMgr);

  bool PIC = false;
  if (Opts.RelocationModel == "static") {
    PIC = false;
  } else if (Opts.RelocationModel == "pic") {
    PIC = true;
  } else {
    assert(Opts.RelocationModel == "dynamic-no-pic" &&
           "Invalid PIC model!");
    PIC = false;
  }

  MOFI->InitMCObjectFileInfo(Triple(Opts.Triple), PIC, Ctx);
  if (Opts.SaveTemporaryLabels)
    Ctx.setAllowTemporaryLabels(false);
  if (Opts.GenDwarfForAssembly)
    Ctx.setGenDwarfForAssembly(true);
  if (!Opts.DwarfDebugFlags.empty())
    Ctx.setDwarfDebugFlags(StringRef(Opts.DwarfDebugFlags));
  if (!Opts.DwarfDebugProducer.empty())
    Ctx.setDwarfDebugProducer(StringRef(Opts.DwarfDebugProducer));
  if (!Opts.DebugCompilationDir.empty())
    Ctx.setCompilationDir(Opts.DebugCompilationDir);
  if (!Opts.DebugPrefixMap.empty())
    for (const auto &KV : Opts.DebugPrefixMap)
      Ctx.addDebugPrefixMapEntry(KV.first, KV.second);
  if (!Opts.MainFileName.empty())
    Ctx.setMainFileName(StringRef(Opts.MainFileName));
  Ctx.setDwarfVersion(Opts.DwarfVersion);

  // Build up the feature string from the target feature list.
  std::string FS;
  if (!Opts.Features.empty()) {
    FS = Opts.Features[0];
    for (unsigned i = 1, e = Opts.Features.size(); i != e; ++i)
      FS += "," + Opts.Features[i];
  }

  std::unique_ptr<MCStreamer> Str;

  std::unique_ptr<MCInstrInfo> MCII(TheTarget->createMCInstrInfo());
  std::unique_ptr<MCSubtargetInfo> STI(
      TheTarget->createMCSubtargetInfo(Opts.Triple, Opts.CPU, FS));

  raw_pwrite_stream *Out = FDOS.get();
  std::unique_ptr<buffer_ostream> BOS;

  // FIXME: There is a bit of code duplication with addPassesToEmitFile.
  if (Opts.OutputType == AssemblerInvocation::FT_Asm) {
    MCInstPrinter *IP = TheTarget->createMCInstPrinter(
        llvm::Triple(Opts.Triple), Opts.OutputAsmVariant, *MAI, *MCII, *MRI);

    std::unique_ptr<MCCodeEmitter> CE;
    if (Opts.ShowEncoding)
      CE.reset(TheTarget->createMCCodeEmitter(*MCII, *MRI, Ctx));
    MCTargetOptions MCOptions;
    std::unique_ptr<MCAsmBackend> MAB(
        TheTarget->createMCAsmBackend(*STI, *MRI, MCOptions));

    auto FOut = llvm::make_unique<formatted_raw_ostream>(*Out);
    Str.reset(TheTarget->createAsmStreamer(
        Ctx, std::move(FOut), /*asmverbose*/ true,
        /*useDwarfDirectory*/ true, IP, std::move(CE), std::move(MAB),
        Opts.ShowInst));
  } else if (Opts.OutputType == AssemblerInvocation::FT_Null) {
    Str.reset(createNullStreamer(Ctx));
  } else {
    assert(Opts.OutputType == AssemblerInvocation::FT_Obj &&
           "Invalid file type!");
    if (!FDOS->supportsSeeking()) {
      BOS = make_unique<buffer_ostream>(*FDOS);
      Out = BOS.get();
    }

    std::unique_ptr<MCCodeEmitter> CE(
        TheTarget->createMCCodeEmitter(*MCII, *MRI, Ctx));
    MCTargetOptions MCOptions;
    std::unique_ptr<MCAsmBackend> MAB(
        TheTarget->createMCAsmBackend(*STI, *MRI, MCOptions));
    std::unique_ptr<MCObjectWriter> OW =
        DwoOS ? MAB->createDwoObjectWriter(*Out, *DwoOS)
              : MAB->createObjectWriter(*Out);

    Triple T(Opts.Triple);
    Str.reset(TheTarget->createMCObjectStreamer(
        T, Ctx, std::move(MAB), std::move(OW), std::move(CE), *STI,
        Opts.RelaxAll, Opts.IncrementalLinkerCompatible,
        /*DWARFMustBeAtTheEnd*/ true));
    Str.get()->InitSections(Opts.NoExecStack);
  }

  // Assembly to object compilation should leverage assembly info.
  Str->setUseAssemblerInfoForParsing(true);

  bool Failed = false;

  std::unique_ptr<MCAsmParser> Parser(
      createMCAsmParser(SrcMgr, Ctx, *Str.get(), *MAI));

  // FIXME: init MCTargetOptions from sanitizer flags here.
  MCTargetOptions Options;
  std::unique_ptr<MCTargetAsmParser> TAP(
      TheTarget->createMCAsmParser(*STI, *Parser, *MCII, Options));
  if (!TAP)
    Failed = Diags.Report(diag::err_target_unknown_triple) << Opts.Triple;

  // Set values for symbols, if any.
  for (auto &S : Opts.SymbolDefs) {
    auto Pair = StringRef(S).split('=');
    auto Sym = Pair.first;
    auto Val = Pair.second;
    int64_t Value;
    // We have already error checked this in the driver.
    Val.getAsInteger(0, Value);
    Ctx.setSymbolValue(Parser->getStreamer(), Sym, Value);
  }

  if (!Failed) {
    Parser->setTargetParser(*TAP.get());
    Failed = Parser->Run(Opts.NoInitialTextSection);
  }

  // Close Streamer first.
  // It might have a reference to the output stream.
  Str.reset();
  // Close the output stream early.
  BOS.reset();
  FDOS.reset();

  // Delete output file if there were errors.
  if (Failed) {
    if (Opts.OutputPath != "-")
      sys::fs::remove(Opts.OutputPath);
    if (!Opts.SplitDwarfFile.empty() && Opts.SplitDwarfFile != "-")
      sys::fs::remove(Opts.SplitDwarfFile);
  }

  return Failed;
}
