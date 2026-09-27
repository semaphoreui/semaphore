import { expect } from 'chai';
import { wheelConsumer, cancelUnconsumedWheel } from '@/lib/wheelGuard';

function scroller(overflowY, { client = 100, scroll = 300, pos = 0 } = {}) {
  const el = document.createElement('div');
  el.style.overflowY = overflowY;
  Object.defineProperty(el, 'clientHeight', { value: client });
  Object.defineProperty(el, 'scrollHeight', { value: scroll });
  el.scrollTop = pos;
  document.body.appendChild(el);
  return el;
}

// jsdom has no WheelEvent; a plain object with the fields the guard reads.
function wheel(target, deltaX, deltaY) {
  const ev = {
    target, deltaX, deltaY, cancelable: true, defaultPrevented: false,
  };
  ev.preventDefault = () => { ev.defaultPrevented = true; };
  return ev;
}

describe('wheelGuard', () => {
  afterEach(() => { document.body.innerHTML = ''; });

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

  it('cancels only wheel events without a consumer', () => {
    const list = scroller('auto');
    const plain = document.createElement('div');
    document.body.appendChild(plain);

    const consumed = wheel(list, 0, 40);
    expect(cancelUnconsumedWheel(consumed)).to.equal(false);
    expect(consumed.defaultPrevented).to.equal(false);

    const stray = wheel(plain, -40, 0);
    expect(cancelUnconsumedWheel(stray)).to.equal(true);
    expect(stray.defaultPrevented).to.equal(true);
  });
});
