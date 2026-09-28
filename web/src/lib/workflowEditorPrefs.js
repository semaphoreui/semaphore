// Browser-local preferences of the workflow editor that App.vue also needs
// before the editor mounts (the navigation drawer must render in its final
// width on the first frame, without a shrink animation).

const SIDE_COLLAPSED_STORAGE_KEY = 'workflowEditor__sideCollapsed';

const EDITOR_PATH = /^\/project\/\d+\/workflows\/(new|\d+\/edit)\/?$/;

export function readSideCollapsed() {
  return localStorage.getItem(SIDE_COLLAPSED_STORAGE_KEY) === '1';
}

export function writeSideCollapsed(collapsed) {
  if (collapsed) {
    localStorage.setItem(SIDE_COLLAPSED_STORAGE_KEY, '1');
  } else {
    localStorage.removeItem(SIDE_COLLAPSED_STORAGE_KEY);
  }
}

export function isWorkflowEditorPath(path) {
  return EDITOR_PATH.test(path || '');
}

// Whether the main navigation should be icon-only for the given route.
export function navMiniForPath(path) {
  return isWorkflowEditorPath(path) && readSideCollapsed();
}
