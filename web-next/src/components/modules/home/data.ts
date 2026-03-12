"use client";

import type { HomeProjectRank } from "@/types";

export type RankSortMode = "deploys" | "successRate" | "failed";

export function rankProjects(projects: HomeProjectRank[], mode: RankSortMode): HomeProjectRank[] {
  return [...projects].sort((left, right) => {
    if (mode === "successRate") {
      if (right.success_rate !== left.success_rate) {
        return right.success_rate - left.success_rate;
      }
      return right.success_count - left.success_count;
    }

    if (mode === "failed") {
      if (right.failed_count !== left.failed_count) {
        return right.failed_count - left.failed_count;
      }
      return right.deploy_count - left.deploy_count;
    }

    if (right.deploy_count !== left.deploy_count) {
      return right.deploy_count - left.deploy_count;
    }
    return right.success_count - left.success_count;
  });
}
