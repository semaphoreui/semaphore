import EventBus from '@/event-bus';

export default async function copyToClipboard(content, message) {
  try {
    await window.navigator.clipboard.writeText(content);
    EventBus.$emit('i-snackbar', {
      color: 'success',
      text: message,
    });
  } catch (e) {
    EventBus.$emit('i-snackbar', {
      color: 'error',
      text: `Can't copy to clipboard: ${e.message}`,
    });
  }
}
