package cmd

// zz_completion_init.go — Deferred completion registration.
//
// This file is named with a "zz_" prefix so that its init() runs AFTER all
// other cmd/*.go init() functions. Go processes init() functions in filename
// order within a package. Cobra's RegisterFlagCompletionFunc requires the
// flag to already exist on the command, so we must wait until every command
// has registered its flags (in their respective init() functions) before
// wiring up completion functions.
//
// Previously this lived in completion.go (starting with 'c'), which ran
// before files like get.go ('g'), set_build_arg.go ('s'), etc., causing
// RegisterFlagCompletionFunc to silently fail for those commands.

func init() {
	// Register custom completion functions for dynamic suggestions.
	// This must run after all command init() functions have registered their flags.
	registerDynamicCompletions()
}
