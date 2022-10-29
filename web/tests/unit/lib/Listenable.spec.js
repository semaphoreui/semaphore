/* eslint-disable symbol-description */
/* eslint-disable no-unused-expressions */
import { expect } from 'chai';
import Listenable from '@/lib/Listenable';

describe('Listenable', () => {
  it('Should add listener', () => {
    const listenable = new Listenable();
    const listenerId = listenable.addListener(() => {});
    expect(listenable.listeners[listenerId]).to.be.ok;
  });

  it('Should remove listener', () => {
    const listenable = new Listenable();
    const listenerId = Symbol();
    listenable.listeners[listenerId] = () => {};
    listenable.removeListener(listenerId);
    expect(listenable.listeners[listenerId]).to.be.undefined;
  });

  it('Should call listener', () => {
    const listenable = new Listenable();
    let d;
    listenable.addListener((data) => {
      d = data;
    });
    listenable.callListeners({
      ok: true,
    });
    expect(d).to.eql({
      ok: true,
    });
  });
});
