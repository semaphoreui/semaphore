import Listenable from '@/lib/Listenable';

export default class Socket extends Listenable {
  constructor(websocketCreator) {
    super();
    this.websocketCreator = websocketCreator;
  }

  start() {
    if (this.ws != null) {
      throw new Error('Websocket already started. Please stop it before starting.');
    }
    this.ws = this.websocketCreator();
    this.ws.onclose = () => {
      if (!this.isRunning()) {
        return;
      }
      this.ws = null;
      setTimeout(() => {
        this.start();
      }, 2000);
    };
    this.ws.onmessage = ({ data }) => {
      this.callListeners(JSON.parse(data));
    };
  }

  isRunning() {
    return this.ws != null;
  }

  stop() {
    if (!this.ws) {
      return;
    }
    this.ws.close();
    delete this.ws;
  }
}
