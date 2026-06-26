import React from "react";

interface LoadingSpinnerProps {
  size?: "sm" | "md" | "lg";
  message?: string;
}

const sizeClasses = {
  sm: "w-5 h-5 border-2",
  md: "w-8 h-8 border-3",
  lg: "w-12 h-12 border-4",
};

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ size = "md", message }) => {
  return (
    <div className="flex flex-col items-center justify-center gap-3" role="status">
      <div
        className={`${sizeClasses[size]} rounded-full border-slate-200 border-t-indigo-600 animate-spin`}
      />
      {message && <p className="text-sm text-slate-500">{message}</p>}
      <span className="sr-only">Loading...</span>
    </div>
  );
};

export function LoadingShell() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50">
      <LoadingSpinner size="lg" message="Loading application..." />
    </div>
  );
}

export function PageLoader() {
  return (
    <div className="flex items-center justify-center py-24">
      <LoadingSpinner size="md" />
    </div>
  );
}
