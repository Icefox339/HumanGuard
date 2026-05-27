type RuntimeConfig = {
  VITE_API_URL?: string;
  PROJECT_VERSION?: string;
};

declare global {
  interface Window {
    __ENV__?: RuntimeConfig;
  }
}

const runtimeConfig = window.__ENV__ ?? {};

export const API_URL = runtimeConfig.VITE_API_URL ?? import.meta.env.VITE_API_URL ?? 'http://localhost:8080';
export const PROJECT_VERSION = runtimeConfig.PROJECT_VERSION ?? import.meta.env.VITE_BUILD_VERSION ?? 'dev';
