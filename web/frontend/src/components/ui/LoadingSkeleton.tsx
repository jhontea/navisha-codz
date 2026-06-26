/**
 * Loading skeleton components for problem list and cards
 */

// ── Skeleton pulse animation ────────────────────────────────────────────────

function SkeletonPulse({ className }: { className?: string }) {
  return (
    <div
      className={`animate-pulse bg-slate-200 dark:bg-slate-700 rounded ${className ?? ""}`}
      aria-hidden="true"
    />
  );
}

// ── Problem Card Skeleton ────────────────────────────────────────────────────

export function ProblemCardSkeleton() {
  return (
    <div className="block bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-3 flex-1 min-w-0">
          <SkeletonPulse className="w-4 h-4 mt-1 shrink-0 rounded-full" />
          <div className="flex-1 min-w-0 space-y-2">
            <SkeletonPulse className="h-5 w-3/4" />
            <div className="flex flex-wrap items-center gap-2">
              <SkeletonPulse className="h-5 w-16 rounded-full" />
              <SkeletonPulse className="h-4 w-20" />
              <SkeletonPulse className="h-4 w-12" />
            </div>
          </div>
        </div>
        <div className="text-right shrink-0 space-y-1">
          <SkeletonPulse className="h-4 w-14 ml-auto" />
          <SkeletonPulse className="h-3 w-10 ml-auto" />
        </div>
      </div>
    </div>
  );
}

// ── Problem List Skeleton ────────────────────────────────────────────────────

export function ProblemListSkeleton({ count = 8 }: { count?: number }) {
  return (
    <div className="space-y-6" aria-label="Loading problems...">
      {/* Filter bar skeleton */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <div className="flex flex-col sm:flex-row gap-3">
          <SkeletonPulse className="flex-1 h-11 rounded-lg" />
          <SkeletonPulse className="w-36 h-11 rounded-lg sm:w-40" />
          <SkeletonPulse className="w-full h-11 rounded-lg sm:w-44" />
          <SkeletonPulse className="w-full h-11 rounded-lg sm:w-36" />
        </div>
      </div>
      {/* Cards skeleton */}
      <div className="space-y-3" role="status">
        {Array.from({ length: count }).map((_, i) => (
          <ProblemCardSkeleton key={i} />
        ))}
        <span className="sr-only">Loading problems...</span>
      </div>
    </div>
  );
}

// ── Filter Bar Skeleton ──────────────────────────────────────────────────────

export function FilterBarSkeleton() {
  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
      <div className="flex flex-col sm:flex-row gap-3">
        <SkeletonPulse className="flex-1 h-11 rounded-lg" />
        <SkeletonPulse className="w-full h-11 rounded-lg sm:w-36" />
        <SkeletonPulse className="w-full h-11 rounded-lg sm:w-44" />
        <SkeletonPulse className="w-full h-11 rounded-lg sm:w-36" />
      </div>
    </div>
  );
}

// ── Detail Page Skeleton ─────────────────────────────────────────────────────

export function ProblemDetailSkeleton() {
  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col space-y-3" role="status">
      {/* Back link */}
      <div className="flex items-center justify-between shrink-0">
        <SkeletonPulse className="h-4 w-32" />
      </div>

      {/* Mobile tab skeleton */}
      <div className="flex lg:hidden gap-1 p-1">
        <SkeletonPulse className="flex-1 h-10 rounded-lg" />
        <SkeletonPulse className="flex-1 h-10 rounded-lg" />
        <SkeletonPulse className="flex-1 h-10 rounded-lg" />
      </div>

      <div className="flex-1 flex flex-col lg:flex-row gap-4 overflow-hidden">
        {/* Left panel skeleton */}
        <div className="flex flex-col space-y-3 lg:w-1/2">
          <SkeletonPulse className="hidden lg:block h-10 rounded-lg" />
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4 sm:p-6 space-y-4 flex-1">
            <SkeletonPulse className="h-8 w-3/4" />
            <SkeletonPulse className="h-5 w-1/2" />
            <SkeletonPulse className="h-4 w-full" />
            <SkeletonPulse className="h-4 w-full" />
            <SkeletonPulse className="h-4 w-2/3" />
            <div className="space-y-2 pt-4">
              <SkeletonPulse className="h-4 w-1/4" />
              <SkeletonPulse className="h-20 w-full rounded-lg" />
            </div>
            <div className="space-y-2">
              <SkeletonPulse className="h-4 w-1/4" />
              <SkeletonPulse className="h-20 w-full rounded-lg" />
            </div>
          </div>
        </div>

        {/* Right panel skeleton */}
        <div className="flex flex-col space-y-3 lg:w-1/2">
          <div className="flex flex-col rounded-lg overflow-hidden border border-slate-700 flex-1">
            <SkeletonPulse className="h-10 w-full rounded-none" />
            <SkeletonPulse className="flex-1 w-full rounded-none" />
          </div>
          <SkeletonPulse className="h-20 w-full rounded-xl" />
        </div>
      </div>
      <span className="sr-only">Loading problem...</span>
    </div>
  );
}
