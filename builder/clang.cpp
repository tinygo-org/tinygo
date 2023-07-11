//go:build byollvm

#include <clang/Basic/DiagnosticOptions.h>
#include <clang/CodeGen/CodeGenAction.h>
#include <clang/Driver/Compilation.h>
#include <clang/Driver/Driver.h>
#include <clang/Frontend/CompilerInstance.h>
#include <clang/Frontend/CompilerInvocation.h>
#include <clang/Frontend/FrontendDiagnostic.h>
#include <clang/Frontend/TextDiagnosticPrinter.h>
#include <clang/FrontendTool/Utils.h>
#include <llvm/ADT/IntrusiveRefCntPtr.h>
#include <llvm/Option/Option.h>
#include <llvm/Support/Host.h>

using namespace llvm;
using namespace clang;

#include "cc1as.h"

// This file provides C wrappers for the builtin tools cc1 and cc1as
// provided by Clang, and calls them as the driver would call them.

extern "C" {

bool tinygo_clang_driver(int argc, char **argv) {
	std::vector<const char*> args(argv, argv + argc);

	// The compiler invocation needs a DiagnosticsEngine so it can report problems
	llvm::IntrusiveRefCntPtr<clang::DiagnosticOptions> DiagOpts = new clang::DiagnosticOptions();
	clang::TextDiagnosticPrinter DiagnosticPrinter(llvm::errs(), &*DiagOpts);
	clang::DiagnosticsEngine Diags(llvm::IntrusiveRefCntPtr<clang::DiagnosticIDs>(new clang::DiagnosticIDs()), &*DiagOpts, &DiagnosticPrinter, false);

	// Create the clang driver
	clang::driver::Driver TheDriver(args[0], llvm::sys::getDefaultTargetTriple(), Diags);

	// Create the set of actions to perform
	std::unique_ptr<clang::driver::Compilation> C(TheDriver.BuildCompilation(args));
	if (!C) {
		return false;
	}
	const clang::driver::JobList &Jobs = C->getJobs();

	// There may be more than one job, for example for .S files
	// (preprocessor + assembler).
	for (auto Cmd : Jobs) {
		// Select the tool: cc1 or cc1as.
		const llvm::opt::ArgStringList &CCArgs = Cmd.getArguments();

		if (strcmp(*CCArgs.data(), "-cc1") == 0) {
			// This is the C frontend.
			// Initialize a compiler invocation object from the clang (-cc1) arguments.
			std::unique_ptr<clang::CompilerInstance> Clang(new clang::CompilerInstance());
			bool success = clang::CompilerInvocation::CreateFromArgs(
					Clang->getInvocation(),
					CCArgs,
					Diags);
			if (!success) {
				return false;
			}

			// Create the actual diagnostics engine.
			Clang->createDiagnostics();
			if (!Clang->hasDiagnostics()) {
				return false;
			}

			// Execute the frontend actions.
			success = ExecuteCompilerInvocation(Clang.get());
			if (!success) {
				return false;
			}

		} else if (strcmp(*CCArgs.data(), "-cc1as") == 0) {
			// This is the assembler frontend. Parse the arguments.
			AssemblerInvocation Asm;
			ArrayRef<const char *> Argv = llvm::ArrayRef<const char*>(CCArgs);
			if (!AssemblerInvocation::CreateFromArgs(Asm, Argv.slice(1), Diags))
				return false;

			// Execute the invocation, unless there were parsing errors.
			bool failed = Diags.hasErrorOccurred() || ExecuteAssembler(Asm, Diags);
			if (failed) {
				return false;
			}

		} else {
			// Unknown tool, print the tool and exit.
			fprintf(stderr, "unknown tool: %s\n", *CCArgs.data());
			return false;
		}
	}

	// Commands executed successfully.
	return true;
}

} // extern "C"
