import { expect } from 'chai';
import PubSub from '@/lib/PubSub';

describe('lib/PubSub', () => {
  it('delivers published data to subscribers of the topic only', () => {
    const bus = new PubSub();
    const a = [];
    const b = [];
    bus.subscribe('a', (data) => a.push(data));
    bus.subscribe('b', (data) => b.push(data));

    bus.publish('a', 1);
    bus.publish('a', 2);

    expect(a).to.deep.equal([1, 2]);
    expect(b).to.deep.equal([]);
  });

  it('calls every subscriber of a topic', () => {
    const bus = new PubSub();
    let calls = 0;
    bus.subscribe('t', () => { calls += 1; });
    bus.subscribe('t', () => { calls += 1; });

    bus.publish('t', null);

    expect(calls).to.equal(2);
  });

  it('ignores publishing to a topic without subscribers', () => {
    const bus = new PubSub();
    expect(() => bus.publish('nobody', 1)).to.not.throw();
  });

  it('stops delivering after unsubscribe', () => {
    const bus = new PubSub();
    const received = [];
    const id = bus.subscribe('t', (data) => received.push(data));
    const otherId = bus.subscribe('t', (data) => received.push(`other:${data}`));

    bus.unsubscribe(id);
    bus.publish('t', 1);

    expect(received).to.deep.equal(['other:1']);
    expect(bus.topics.t.hasListeners()).to.equal(true);

    bus.unsubscribe(otherId);
    expect(bus.topics.t.hasListeners()).to.equal(false);
  });

  it('ignores unsubscribe with an unknown id', () => {
    const bus = new PubSub();
    bus.subscribe('t', () => {});
    expect(() => bus.unsubscribe(Symbol('unknown'))).to.not.throw();
  });
});
