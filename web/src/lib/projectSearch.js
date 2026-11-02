/**
 * The project selector shows its search field only above this many projects:
 * a short list is quicker to scan than to search.
 */
export const PROJECT_SEARCH_MIN_PROJECTS = 5;

/**
 * Projects whose name contains the query, ignoring case and surrounding
 * spaces. An empty query returns all projects.
 *
 * @param {Array<{name: string}>|null|undefined} projects
 * @param {string|null|undefined} query
 * @returns {Array<{name: string}>}
 */
export function filterProjects(projects, query) {
  const list = projects || [];
  const q = (query || '').trim().toLocaleLowerCase();

  if (q === '') {
    return list;
  }

  return list.filter((p) => (p.name || '').toLocaleLowerCase().includes(q));
}
