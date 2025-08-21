package i18n

// Message IDs for consistent reference across the application
const (
	// Command descriptions
	MsgCmdRootShort       = "cmd.root.short"
	MsgCmdRootLong        = "cmd.root.long"
	MsgCmdReplaceShort    = "cmd.replace.short"
	MsgCmdReplaceLong     = "cmd.replace.long"
	MsgCmdCreateShort     = "cmd.create.short"
	MsgCmdCreateLong      = "cmd.create.long"
	MsgCmdTemplateShort   = "cmd.template.short"
	MsgCmdTemplateLong    = "cmd.template.long"
	MsgCmdGenerateShort   = "cmd.generate.short"
	MsgCmdGenerateLong    = "cmd.generate.long"
	MsgCmdVersionShort    = "cmd.version.short"
	MsgCmdVersionLong     = "cmd.version.long"

	// Flag descriptions
	MsgFlagRules          = "flag.rules"
	MsgFlagPath           = "flag.path"
	MsgFlagDryRun         = "flag.dryrun"
	MsgFlagBackup         = "flag.backup"
	MsgFlagRecursive      = "flag.recursive"
	MsgFlagFrom           = "flag.from"
	MsgFlagTemplate       = "flag.template"
	MsgFlagOutput         = "flag.output"
	MsgFlagFormat         = "flag.format"
	MsgFlagForce          = "flag.force"
	MsgFlagValues         = "flag.values"
	MsgFlagSet            = "flag.set"
	MsgFlagType           = "flag.type"
	MsgFlagPrompt         = "flag.prompt"
	MsgFlagLang           = "flag.lang"

	// Success messages
	MsgSuccessCreated     = "success.created"
	MsgSuccessReplaced    = "success.replaced"
	MsgSuccessBackup      = "success.backup"
	MsgSuccessProcessed   = "success.processed"

	// Progress messages
	MsgProgressConverting = "progress.converting"
	MsgProgressProcessing = "progress.processing"
	MsgProgressRules      = "progress.rules"
	MsgProgressDryRun     = "progress.dryrun"

	// Warning messages
	MsgWarningNoValues    = "warning.no_values"
	MsgWarningTemplate    = "warning.template_not_impl"
	MsgWarningNoRules     = "warning.no_rules"

	// Error messages
	MsgErrorFileNotFound  = "error.file_not_found"
	MsgErrorFileExists    = "error.file_exists"
	MsgErrorInvalidFormat = "error.invalid_format"
	MsgErrorInvalidSet    = "error.invalid_set"
	MsgErrorUnsupported   = "error.unsupported"
	MsgErrorLoadRules     = "error.load_rules"
	MsgErrorLoadValues    = "error.load_values"
	MsgErrorProcess       = "error.process"
	MsgErrorValidate      = "error.validate"
	MsgErrorConversion    = "error.conversion"
	MsgErrorNotImpl       = "error.not_implemented"
	MsgErrorParseYAML     = "error.parse_yaml"
	MsgErrorParseJSON     = "error.parse_json"
	MsgErrorParseFile     = "error.parse_file"
	MsgErrorCreateBackup  = "error.create_backup"
	MsgErrorAccessPath    = "error.access_path"

	// Summary messages
	MsgSummaryTotal       = "summary.total"
	MsgSummarySuccess     = "summary.success"
	MsgSummaryFailed      = "summary.failed"
	MsgSummarySkipped     = "summary.skipped"
	MsgSummaryResults     = "summary.results"
)