import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { execSync } from 'node:child_process';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const getBuildVersion = () => {
  try {
    const commits = execSync('git rev-list --count HEAD', { cwd: __dirname }).toString().trim();
    const hash = execSync('git rev-parse --short HEAD', { cwd: __dirname }).toString().trim();
    return `v${commits}-${hash}`;
  } catch {
    return `dev-${Date.now()}`;
  }
};

export default defineConfig({
  plugins: [react()],
  define: {
    'import.meta.env.VITE_BUILD_VERSION': JSON.stringify(getBuildVersion())
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  server: {
    port: 5173
  }
});
