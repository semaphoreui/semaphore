import EventBus from '@/event-bus';

export default async function copyToClipboard(text) {
  try {
    await window.navigator.clipboard.writeText(text);
    EventBus.$emit('i-snackbar', {
      color: 'success',
      text: 'The license key has been copied to the clipboard.',
    });
  } catch (e) {
    EventBus.$emit('i-snackbar', {
      color: 'error',
      text: `Can't copy the license key: ${e.message}`,
    });
  }
}
