/**
 * Shared metadata for project alerts. Channel descriptions (type, fields,
 * default events, default body) come from the API; this file only holds the
 * presentation of the fixed event set.
 */

export const ALERT_EVENTS = [
  {
    value: 'success',
    icon: 'mdi-check-circle-outline',
    color: 'success',
    label: 'alertEventSuccess',
  },
  {
    value: 'error',
    icon: 'mdi-alert-circle-outline',
    color: 'error',
    label: 'alertEventError',
  },
  {
    value: 'waiting_confirmation',
    icon: 'mdi-account-clock-outline',
    color: 'warning',
    label: 'alertEventWaitingConfirmation',
  },
];

export const ALERT_FIELD_LABELS = {
  chat_id: 'telegramChatId',
  thread_id: 'telegramThreadId',
  url: 'alertWebhookUrl',
  token: 'alertToken',
  recipients: 'alertRecipients',
};

/**
 * Returns the events an alert listens to: its own list or, when empty, the
 * channel defaults.
 */
export function effectiveAlertEvents(alert, channel) {
  if (alert && Array.isArray(alert.events) && alert.events.length > 0) {
    return alert.events;
  }
  return channel ? channel.default_events || [] : [];
}

export function findChannel(channels, type) {
  return (channels || []).find((c) => c.type === type) || null;
}
