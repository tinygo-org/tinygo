// Source: https://github.com/llvm/llvm-project/blob/main/clang/tools/driver/cc1as_main.cpp
// See cc1as.cpp for details.

//===-- cc1as.h - Clang Assembler  ----------------------------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//
//
// This is the entry point to the clang -cc1as functionality, which implements
// the direct interface to the LLVM MC based assembler.
//
//===----------------------------------------------------------------------===//

/// Helper class for representing a single invocation of the assembler.
struct AssemblerInvocation {
  /// @name Target Options
  /// @{

  /// The name of the target triple to assemble for.
  std::string Triple;

  /// If given, the name of the target CPU to determine which instructions
  /// are legal.
  std::string CPU;

  /// The list of target specific features to enable or disable -- this should
  /// be a list of strings starting with '+' or '-'.
  std::vector<std::string> Features;

  /// The list of symbol definitions.
  std::vector<std::string> SymbolDefs;

  /// @}
  /// @name Language Options
  /// @{

  std::vector<std::string> IncludePaths;
  unsigned NoInitialTextSection : 1;
  unsigned SaveTemporaryLabels : 1;
  unsigned GenDwarfForAssembly : 1;
  unsigned RelaxELFRelocations : 1;
  unsigned Dwarf64 : 1;
  unsigned DwarfVersion;
  std::string DwarfDebugFlags;
  std::string DwarfDebugProducer;
  std::string DebugCompilationDir;
  llvm::SmallVector<std::pair<std::string, std::string>, 0> DebugPrefixMap;
  llvm::DebugCompressionType CompressDebugSections =
      llvm::DebugCompressionType::None;
  std::string MainFileName;
  std::string SplitDwarfOutput;

  /// @}
  /// @name Frontend Options
  /// @{

  std::string InputFile;
  std::vector<std::string> LLVMArgs;
  std::string OutputPath;
  enum FileType {
    FT_Asm,  ///< Assembly (.s) output, transliterate mode.
    FT_Null, ///< No output, for timing purposes.
    FT_Obj   ///< Object file output.
  };
  FileType OutputType;
  unsigned ShowHelp : 1;
  unsigned ShowVersion : 1;

  /// @}
  /// @name Transliterate Options
  /// @{

  unsigned OutputAsmVariant;
  unsigned ShowEncoding : 1;
  unsigned ShowInst : 1;

  /// @}
  /// @name Assembler Options
  /// @{

  unsigned RelaxAll : 1;
  unsigned NoExecStack : 1;
  unsigned FatalWarnings : 1;
  unsigned NoWarn : 1;
  unsigned NoTypeCheck : 1;
  unsigned IncrementalLinkerCompatible : 1;
  unsigned EmbedBitcode : 1;

  /// Whether to emit DWARF unwind info.
  EmitDwarfUnwindType EmitDwarfUnwind;

  // Whether to emit compact-unwind for non-canonical entries.
  // Note: may be overridden by other constraints.
  unsigned EmitCompactUnwindNonCanonical : 1;

  /// The name of the relocation model to use.
  std::string RelocationModel;

  /// The ABI targeted by the backend. Specified using -target-abi. Empty
  /// otherwise.
  std::string TargetABI;

  /// Darwin target variant triple, the variant of the deployment target
  /// for which the code is being compiled.
  std::optional<llvm::Triple> DarwinTargetVariantTriple;

  /// The version of the darwin target variant SDK which was used during the
  /// compilation
  llvm::VersionTuple DarwinTargetVariantSDKVersion;

  /// The name of a file to use with \c .secure_log_unique directives.
  std::string AsSecureLogFile;
  /// @}

public:
  AssemblerInvocation() {
    Triple = "";
    NoInitialTextSection = 0;
    InputFile = "-";
    OutputPath = "-";
    OutputType = FT_Asm;
    OutputAsmVariant = 0;
    ShowInst = 0;
    ShowEncoding = 0;
    RelaxAll = 0;
    NoExecStack = 0;
    FatalWarnings = 0;
    NoWarn = 0;
    NoTypeCheck = 0;
    IncrementalLinkerCompatible = 0;
    Dwarf64 = 0;
    DwarfVersion = 0;
    EmbedBitcode = 0;
    EmitDwarfUnwind = EmitDwarfUnwindType::Default;
    EmitCompactUnwindNonCanonical = false;
  }

  static bool CreateFromArgs(AssemblerInvocation &Res,
                             ArrayRef<const char *> Argv,
                             DiagnosticsEngine &Diags);
};

bool ExecuteAssembler(AssemblerInvocation &Opts, DiagnosticsEngine &Diags);
