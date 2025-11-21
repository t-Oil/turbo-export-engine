import * as path from 'path';
import * as os from 'os';

export interface PlatformInfo {
  platform: string;
  arch: string;
  binaryName: string;
  binaryPath: string;
}

export function detectPlatform(): PlatformInfo {
  const platform = os.platform();
  const arch = os.arch();

  let binaryName: string;

  switch (platform) {
    case 'darwin':
      binaryName = 'export-engine-macos';
      break;
    case 'linux':
      binaryName = 'export-engine-linux';
      break;
    case 'win32':
      binaryName = 'export-engine-win.exe';
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }

  // Binary path relative to node_modules
  const binaryPath = path.join(__dirname, '..', 'bin', binaryName);

  return {
    platform,
    arch,
    binaryName,
    binaryPath,
  };
}

export function validatePlatform(): void {
  const { platform, arch } = detectPlatform();

  const supported = [
    { platform: 'darwin', arch: 'x64' },
    { platform: 'darwin', arch: 'arm64' },
    { platform: 'linux', arch: 'x64' },
    { platform: 'win32', arch: 'x64' },
  ];

  const isSupported = supported.some(
    (s) => s.platform === platform && s.arch === arch
  );

  if (!isSupported) {
    throw new Error(
      `Unsupported platform: ${platform} ${arch}. Supported: darwin/linux/win32 on x64/arm64`
    );
  }
}
