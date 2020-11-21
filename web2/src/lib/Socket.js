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
      setTimeout(() => {
        this.start();
      }, 2000);
    };
    this.ws.onmessage = ({ data }) => {
      try {
        this.callListeners(JSON.parse(data));
      } catch (e) {
        console.error(e);
      }
    };
  }

  stop() {
    this.ws.close();
    delete this.ws;
  }

  addListener(callback) {
    const isFirstListener = !this.hasListeners();
    const listenerId = super.addListener(callback);
    if (isFirstListener) {
      this.start();
    }
    return listenerId;
  }

  removeListener(id) {
    super.removeListener(id);
    if (!this.hasListeners()) {
      this.stop();
    }
  }
}
