import * as vscode from 'vscode';
import { exec, execSync } from 'child_process';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
} from 'vscode-languageclient/node';

let client: LanguageClient | undefined;

const INSTALL_COMMAND = 'go install github.com/grindlemire/go-tui/cmd/tui@latest';

function binaryExists(command: string): boolean {
    try {
        const whichCmd = process.platform === 'win32' ? 'where' : 'which';
        execSync(`${whichCmd} ${command}`, { stdio: 'ignore' });
        return true;
    } catch {
        return false;
    }
}

async function installBinary(): Promise<boolean> {
    return new Promise((resolve) => {
        vscode.window.withProgress(
            {
                location: vscode.ProgressLocation.Notification,
                title: 'Installing GSX language server...',
                cancellable: false,
            },
            async () => {
                return new Promise<void>((progressResolve) => {
                    exec(INSTALL_COMMAND, (error, stdout, stderr) => {
                        if (error) {
                            vscode.window.showErrorMessage(
                                `Failed to install tui: ${stderr || error.message}`
                            );
                            resolve(false);
                        } else {
                            vscode.window.showInformationMessage(
                                'GSX language server installed successfully. Reload window to activate.'
                            );
                            resolve(true);
                        }
                        progressResolve();
                    });
                });
            }
        );
    });
}

async function promptInstall(): Promise<boolean> {
    const selection = await vscode.window.showWarningMessage(
        'GSX language server not found.',
        'Install',
        'Disable LSP'
    );

    if (selection === 'Install') {
        return installBinary();
    } else if (selection === 'Disable LSP') {
        const config = vscode.workspace.getConfiguration('gsx.lsp');
        await config.update('enabled', false, vscode.ConfigurationTarget.Global);
        vscode.window.showInformationMessage(
            'GSX LSP disabled. Re-enable in settings with gsx.lsp.enabled'
        );
    }

    return false;
}

export async function activate(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('gsx.lsp');
    const enabled = config.get<boolean>('enabled', true);

    if (!enabled) {
        return;
    }

    const tuiPath = config.get<string>('path', 'tui');

    if (!binaryExists(tuiPath)) {
        const installed = await promptInstall();
        if (!installed) {
            return;
        }
    }

    const logPath = config.get<string>('logPath', '');
    const args = ['lsp'];
    if (logPath) {
        args.push('-log', logPath);
    }

    const serverOptions: ServerOptions = {
        command: tuiPath,
        args: args,
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'gsx' }],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.gsx'),
        },
    };

    client = new LanguageClient(
        'gsxLanguageServer',
        'GSX Language Server',
        serverOptions,
        clientOptions
    );

    client.start();
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
