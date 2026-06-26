import React from "react";
import { useQuery } from "react-query";
import { problemApi } from "../services/api";
import { ProblemList } from "../components/ProblemList";
import { PageLoader } from "../components/ui/LoadingSpinner";
import { queryKeys } from "../hooks/useQueries";

export function ProblemsPage() {
  const [page, setPage] = React.useState(1);
  const [search, setSearch] = React.useState("");
  const [difficulty, setDifficulty] = React.useState("");
  const [category, setCategory] = React.useState("");

  const { data, isLoading, error } = useQuery(
    queryKeys.problems.list({ page, page_size: 20, search, difficulty, category }),
    () =>
      problemApi
        .list({ page, page_size: 20, search, difficulty, category })
        .then((r) => r.data),
    { staleTime: 60000 }
  );

  if (isLoading) return <PageLoader />;
  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500 dark:text-red-400">Failed to load problems. Please try again.</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Problems</h1>
        <p className="text-slate-500 dark:text-slate-400 mt-1">
          {data?.total ?? 0} problems to sharpen your skills
        </p>
      </div>
      <ProblemList
        problems={data?.items ?? []}
        total={data?.total ?? 0}
        page={page}
        pageSize={20}
        onPageChange={setPage}
      />
    </div>
  );
}
