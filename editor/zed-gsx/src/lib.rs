use zed_extension_api as zed;

struct GsxExtension;

impl zed::Extension for GsxExtension {
    fn new() -> Self {
        GsxExtension
    }

    fn language_server_command(
        &mut self,
        _language_server_id: &zed::LanguageServerId,
        worktree: &zed::Worktree,
    ) -> Result<zed::Command, String> {
        let path = worktree
            .which("tui")
            .ok_or_else(|| "tui binary not found in PATH".to_string())?;

        Ok(zed::Command {
            command: path,
            args: vec!["lsp".to_string()],
            env: Default::default(),
        })
    }
}

zed::register_extension!(GsxExtension);
