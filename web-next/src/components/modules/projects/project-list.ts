import type { Project } from "@/types";

export function mergeProjectIntoList(projects: Project[], project: Project): Project[] {
  return [project, ...projects.filter((item) => item.id !== project.id)];
}
