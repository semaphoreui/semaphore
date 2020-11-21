import Listenable from '@/lib/Listenable';

export default class PubSub {
  constructor() {
    this.topics = {};
  }

  subscribe(topic, callback) {
    if (this.topics[topic] == null) {
      this.topics[topic] = new Listenable();
    }
    return this.topics[topic].addListener(callback);
  }

  unsubscribe(id) {
    // eslint-disable-next-line no-restricted-syntax
    for (const topic in this.topics) {
      if (this.topics[topic].removeListener(id)) {
        break;
      }
    }
  }

  publish(topic, data) {
    if (!this.topics[topic]) {
      return;
    }
    this.topics[topic].callListeners(data);
  }
}
