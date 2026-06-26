import React from "react";
import { Download, RefreshCw, X, WifiOff } from "lucide-react";
import { usePWAInstall, useServiceWorker } from "../hooks/usePWA";

// PWA Install Prompt Component
export function PWAInstallPrompt() {
  const { isInstallable, isInstalled, install } = usePWAInstall();
  const [dismissed, setDismissed] = React.useState(false);

  if (!isInstallable || isInstalled || dismissed) return null;

  return (
    <div
      className="fixed bottom-4 left-4 right-4 sm:left-auto sm:right-4 sm:w-96 bg-white rounded-xl shadow-xl border border-slate-200 p-4 z-50 animate-slideUp"
      role="dialog"
      aria-label="Install app"
    >
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0 w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center">
          <Download className="w-5 h-5 text-indigo-600" />
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-semibold text-slate-900">Install CodeChallenge</h3>
          <p className="text-xs text-slate-500 mt-1">
            Install the app for offline access and a better experience.
          </p>
          <div className="flex items-center gap-2 mt-3">
            <button
              onClick={install}
              className="px-3 py-1.5 text-xs font-medium bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
            >
              Install
            </button>
            <button
              onClick={() => setDismissed(true)}
              className="px-3 py-1.5 text-xs font-medium text-slate-500 hover:text-slate-700 transition-colors"
            >
              Not now
            </button>
          </div>
        </div>
        <button
          onClick={() => setDismissed(true)}
          className="flex-shrink-0 p-1 text-slate-400 hover:text-slate-600"
          aria-label="Dismiss"
        >
          <X className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}

// Update Available Banner
export function PWAUpdateBanner() {
  const { hasUpdate, skipWaiting } = useServiceWorker();

  if (!hasUpdate) return null;

  return (
    <div
      className="fixed top-0 left-0 right-0 bg-indigo-600 text-white px-4 py-3 z-50 flex items-center justify-between"
      role="alert"
    >
      <div className="flex items-center gap-2">
        <RefreshCw className="w-4 h-4" />
        <span className="text-sm font-medium">A new version is available!</span>
      </div>
      <button
        onClick={skipWaiting}
        className="px-3 py-1 text-sm font-medium bg-white text-indigo-600 rounded-lg hover:bg-indigo-50 transition-colors"
      >
        Update
      </button>
    </div>
  );
}

// Offline Indicator
export function OfflineIndicator() {
  const { isOnline } = useServiceWorker();

  if (isOnline) return null;

  return (
    <div
      className="fixed top-0 left-0 right-0 bg-amber-500 text-white px-4 py-2 z-50 flex items-center justify-center gap-2"
      role="alert"
    >
      <WifiOff className="w-4 h-4" />
      <span className="text-sm font-medium">You are offline. Some features may be limited.</span>
    </div>
  );
}
