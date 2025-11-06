import EventBus from '@/event-bus';

export default async function copyToClipboard(content, message) {
  try {
    const el = document.createElement('textarea');
    el.value = this.link_url;
    el.setAttribute('readonly', '');
    el.style.position = 'absolute';
    el.style.left = '-9999px';
    document.body.appendChild(el);
    const selected = document.getSelection().rangeCount > 0
      ? document.getSelection().getRangeAt(0) : false;
    el.select();
    document.execCommand('copy');
    document.body.removeChild(el);
    if (selected) {
      document.getSelection().removeAllRanges();
      document.getSelection().addRange(selected);
    }

    const successful = document.execCommand('copy');
    // document.body.removeChild(textArea);

    if (!successful) {
      throw new Error('Fallback copy failed');
    }

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
