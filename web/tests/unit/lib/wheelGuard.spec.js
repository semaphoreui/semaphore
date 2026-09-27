import { expect } from 'chai';
import {
  wheelConsumer, wheelPixels, scrollByWheel, routeWheel,
} from '@/lib/wheelGuard';

function scroller(overflowY, { client = 100, scroll = 300, pos = 0 } = {}) {
  const el = document.createElement('div');
  el.style.overflowY = overflowY;
  Object.defineProperty(el, 'clientHeight', { value: client });
  Object.defineProperty(el, 'scrollHeight', { value: scroll });
  el.scrollTop = pos;
  document.body.appendChild(el);
  return el;
}

// jsdom has no WheelEvent; a plain object with the fields the router reads.
function wheel(deltaX, deltaY, extra = {}) {
  const ev = {
    deltaX,
    deltaY,
    deltaMode: 0,
    clientX: 10,
    clientY: 10,
    cancelable: true,
    defaultPrevented: false,
    stopped: false,
    ...extra,
  };
  ev.preventDefault = () => { ev.defaultPrevented = true; };
  ev.stopImmediatePropagation = () => { ev.stopped = true; };
  return ev;
}

describe('wheelGuard', () => {
  let underPointer = null;
  before(() => {
    document.elementFromPoint = () => underPointer;
    if (!Element.prototype.scrollBy) {
      Element.prototype.scrollBy = function scrollBy(x, y) {
        this.scrollLeft += x;
        this.scrollTop += y;
      };
    }
  });
  afterEach(() => { document.body.innerHTML = ''; underPointer = null; });

  it('finds a scrollable ancestor that can move in the wheel direction', () => {
    const list = scroller('auto');
    const item = document.createElement('span');
    list.appendChild(item);
    expect(wheelConsumer(item, 0, 40)).to.equal(list);
  });

  it('ignores a scroller already at its end', () => {
    const list = scroller('auto', { pos: 200 });
    expect(wheelConsumer(list, 0, 40)).to.equal(null);
    expect(wheelConsumer(list, 0, -40)).to.equal(list);
  });

  it('ignores overflow hidden and non-overflowing content', () => {
    expect(wheelConsumer(scroller('hidden'), 0, 40)).to.equal(null);
    expect(wheelConsumer(scroller('auto', { scroll: 100 }), 0, 40)).to.equal(null);
  });

  it('uses the dominant axis only', () => {
    const list = scroller('auto');
    // Mostly horizontal swipe over a vertical list: nothing consumes it.
    expect(wheelConsumer(list, -40, 5)).to.equal(null);
  });

  it('converts line deltas to pixels', () => {
    expect(wheelPixels(wheel(0, 3, { deltaMode: 1 }), 'y')).to.equal(48);
    expect(wheelPixels(wheel(0, 3), 'y')).to.equal(3);
  });

  it('scrolls an element by the wheel delta', () => {
    const list = scroller('auto', { pos: 10 });
    scrollByWheel(list, wheel(0, 40));
    expect(list.scrollTop).to.equal(50);
  });

  describe('routeWheel', () => {
    it('sends the event to the graph when the pointer is over it', () => {
      const graph = document.createElement('div');
      document.body.appendChild(graph);
      underPointer = graph;
      const ev = wheel(-40, 0);
      let received = null;
      const result = routeWheel(ev, {
        isGraph: (el) => el === graph,
        onGraph: (e) => { received = e; },
      });
      expect(result).to.equal('graph');
      expect(received).to.equal(ev);
      expect(ev.defaultPrevented).to.equal(true);
      expect(ev.stopped).to.equal(true);
    });

    it('routes by pointer position, not by event target', () => {
      const graph = document.createElement('div');
      document.body.appendChild(graph);
      underPointer = graph;
      const ev = wheel(-40, 0, { target: scroller('auto') });
      expect(routeWheel(ev, { isGraph: (el) => el === graph, onGraph: () => {} })).to.equal('graph');
    });

    it('scrolls a scrollable element under the pointer and cancels the event', () => {
      const list = scroller('auto');
      underPointer = list;
      const ev = wheel(0, 40);
      expect(routeWheel(ev, { isGraph: () => false, onGraph: () => {} })).to.equal('scrolled');
      expect(list.scrollTop).to.equal(40);
      expect(ev.defaultPrevented).to.equal(true);
      expect(ev.stopped).to.equal(false);
    });

    it('drops and cancels an event nothing can consume', () => {
      const plain = document.createElement('div');
      document.body.appendChild(plain);
      underPointer = plain;
      const ev = wheel(-40, 0);
      expect(routeWheel(ev, { isGraph: () => false, onGraph: () => {} })).to.equal('dropped');
      expect(ev.defaultPrevented).to.equal(true);
    });

    it('leaves non-cancelable events alone', () => {
      const ev = wheel(-40, 0, { cancelable: false });
      expect(routeWheel(ev, { isGraph: () => true, onGraph: () => {} })).to.equal('ignored');
      expect(ev.defaultPrevented).to.equal(false);
    });
  });
});
