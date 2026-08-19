import Listenable from '@/lib/Listenable';
import FakeWebSocket from './FakeWebSocket';

export default class Socket extends Listenable {
  constructor(websocketCreator) {
    super();
    this.websocketCreator = websocketCreator;
    this.sessionActive = false;
    this.wantStart = false;
  }

  /**
   * Sets whether the user session is active.
   * If session becomes active and start() was previously called, creates real WebSocket.
   * If session becomes inactive, stops real WebSocket and switches to fake one.
   * @param {boolean} isActive - Whether the user session is active
   */
  setSessionActive(isActive) {
    const wasActive = this.sessionActive;
    this.sessionActive = isActive;

    if (isActive && !wasActive && this.wantStart) {
      // Session became active and we wanted to start - create real socket
      this.startRealSocket();
    } else if (!isActive && wasActive && this.ws) {
      // Session became inactive - stop real socket
      this.stop();
    }
  }

  /**
   * Internal method to start a real WebSocket connection.
   * Only called when session is active.
   */
  startRealSocket() {
    if (this.ws != null) {
      return;
    }
    this.ws = this.websocketCreator();
    this.ws.onclose = () => {
      if (!this.isRunning()) {
        return;
      }
      this.ws = null;
      setTimeout(() => {
        if (this.sessionActive) {
          this.startRealSocket();
        }
      }, 2000);
    };
    this.ws.onmessage = ({ data }) => {
      this.callListeners(JSON.parse(data));
    };
  }

  start() {
    this.wantStart = true;

    if (this.ws != null) {
      return; // Already running (real or fake)
    }

    if (this.sessionActive) {
      // Session is active, create real WebSocket
      this.startRealSocket();
    } else {
      // No session, create fake WebSocket to avoid errors
      this.ws = new FakeWebSocket();
    }
  }

  isRunning() {
    return this.ws != null;
  }

  stop() {
    this.wantStart = false;
    if (!this.ws) {
      return;
    }
    this.ws.close();
    delete this.ws;
  }
}
