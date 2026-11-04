/**
 * The project selector shows its search field only above this many projects:
 * a short list is quicker to scan than to search.
 */
export const PROJECT_SEARCH_MIN_PROJECTS = 5;

// Lower case without diacritics, so that "cafe" finds "Café": NFD splits a
// letter from its combining marks (U+0300-U+036F), which are then dropped.
function fold(text) {
  return (text || '')
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLocaleLowerCase();
}

/**
 * Projects whose name contains the query, ignoring case, diacritics and
 * surrounding spaces. An empty query returns all projects.
 *
 * @param {Array<{name: string}>|null|undefined} projects
 * @param {string|null|undefined} query
 * @returns {Array<{name: string}>}
 */
export function filterProjects(projects, query) {
  const list = projects || [];
  const q = fold(query).trim();

  if (q === '') {
    return list;
  }

  return list.filter((p) => fold(p.name).includes(q));
}
