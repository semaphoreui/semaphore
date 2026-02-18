/**
 * A fake WebSocket that mimics the WebSocket interface but does nothing.
 * Used when there's no active user session to prevent console errors.
 */
class FakeWebSocket {
  constructor() {
    this.readyState = WebSocket.OPEN;
  }

  // eslint-disable-next-line class-methods-use-this
  close() {
    // Do nothing
  }

  // eslint-disable-next-line class-methods-use-this
  send() {
    // Do nothing
  }
}

export default FakeWebSocket;
