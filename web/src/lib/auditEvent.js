// Display rules for audit events, the SIEM envelope returned by /api/audit/events.

export function eventKind(event) {
  return `${event.category}.${event.event_code}/${event.action}`;
}

function keyOf(prefix, value) {
  return `${prefix}_${value.replace(/[./]/g, '_')}`;
}

// Titles exist in en.js only, so lookups check English to work in every locale.
function translate(i18n, key, fallback, values) {
  return i18n.te(key, 'en') ? i18n.t(key, values) : fallback;
}

export function eventTitle(event, i18n) {
  const key = keyOf('audit_kind', eventKind(event));
  const success = translate(i18n, key, eventKind(event));
  return event.outcome === 'failure' ? translate(i18n, `${key}_failure`, success) : success;
}

export function reasonLabel(reason, i18n) {
  return reason ? translate(i18n, keyOf('audit_reason', reason), reason) : '';
}

export function targetTypeLabel(type, i18n) {
  return type ? translate(i18n, keyOf('audit_target', type), type) : '';
}

export function actorLabel(actor, i18n) {
  const name = actor.name || actor.id || '';
  switch (actor.type) {
    case 'user':
      return actor.auth === 'api_token' ? i18n.t('audit_actor_api_token', { name }) : name;
    case 'runner':
      return i18n.t('audit_actor_runner', { name });
    case 'system':
      return i18n.t('audit_actor_system', { name });
    case 'integration':
      return i18n.t('audit_actor_integration', { name });
    default:
      return name || i18n.t('audit_actor_anonymous');
  }
}

export function actorIcon(actor) {
  if (actor.type === 'user') {
    return actor.auth === 'api_token' ? 'mdi-api' : 'mdi-account';
  }
  return {
    runner: 'mdi-cogs',
    system: 'mdi-server',
    integration: 'mdi-webhook',
  }[actor.type] || 'mdi-incognito';
}

// Nested metadata becomes dotted keys, lists become comma-separated text.
export function metadataRows(metadata, prefix = '') {
  return Object.entries(metadata || {}).flatMap(([key, value]) => {
    const path = prefix ? `${prefix}.${key}` : key;
    if (Array.isArray(value)) {
      return [{ key: path, value: value.join(', ') }];
    }
    if (value !== null && typeof value === 'object') {
      return metadataRows(value, path);
    }
    return [{ key: path, value: String(value) }];
  });
}

// Empty filters are left out so the request carries only what the user chose.
export function filterParams(filters) {
  return Object.fromEntries(Object.entries(filters).filter(([, value]) => (
    Array.isArray(value) ? value.length > 0 : value !== null && value !== undefined && value !== ''
  )));
}
