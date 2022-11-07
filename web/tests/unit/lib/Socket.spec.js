/* eslint-disable symbol-description */
/* eslint-disable no-unused-expressions */
import { expect } from 'chai';
import Socket from '@/lib/Socket';

describe('Socket', () => {
  it('Should add listener', () => {
    const socket = new Socket(() => ({
      close: () => {},
    }));
    expect(socket.ws).to.be.ok;
  });
});
