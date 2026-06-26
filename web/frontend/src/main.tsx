import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import { PWAInstallPrompt, PWAUpdateBanner, OfflineIndicator } from "./components/PWAComponents";
import "./index.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
    <PWAInstallPrompt />
    <PWAUpdateBanner />
    <OfflineIndicator />
  </React.StrictMode>
);
