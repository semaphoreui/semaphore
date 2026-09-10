import { expect } from 'chai';
import Socket from '@/lib/Socket';
import FakeWebSocket from '@/lib/FakeWebSocket';

function createFakeCreator() {
  const created = [];
  const creator = () => {
    const ws = {
      closed: false,
      close() {
        this.closed = true;
      },
    };
    created.push(ws);
    return ws;
  };
  return { creator, created };
}

describe('Socket', () => {
  it('uses a fake WebSocket when started without an active session', () => {
    const { creator, created } = createFakeCreator();
    const socket = new Socket(creator);

    socket.start();

    expect(socket.isRunning()).to.equal(true);
    expect(socket.ws).to.be.instanceOf(FakeWebSocket);
    expect(created).to.have.lengthOf(0);
  });

  it('creates a real WebSocket when the session is active', () => {
    const { creator, created } = createFakeCreator();
    const socket = new Socket(creator);

    socket.setSessionActive(true);
    socket.start();

    expect(created).to.have.lengthOf(1);
    expect(socket.ws).to.equal(created[0]);
  });

  it('switches from fake to real WebSocket when the session becomes active', () => {
    const { creator, created } = createFakeCreator();
    const socket = new Socket(creator);

    socket.start();
    expect(socket.ws).to.be.instanceOf(FakeWebSocket);

    socket.setSessionActive(true);

    expect(created).to.have.lengthOf(1);
    expect(socket.ws).to.equal(created[0]);
  });

  it('dispatches parsed messages to listeners', () => {
    const { creator, created } = createFakeCreator();
    const socket = new Socket(creator);
    const received = [];
    socket.addListener((data) => received.push(data));

    socket.setSessionActive(true);
    socket.start();
    created[0].onmessage({ data: JSON.stringify({ type: 'update', id: 7 }) });

    expect(received).to.deep.equal([{ type: 'update', id: 7 }]);
  });

  it('closes the WebSocket on stop', () => {
    const { creator, created } = createFakeCreator();
    const socket = new Socket(creator);

    socket.setSessionActive(true);
    socket.start();
    socket.stop();

    expect(created[0].closed).to.equal(true);
    expect(socket.isRunning()).to.equal(false);
  });

  it('closes the WebSocket when the session becomes inactive', () => {
    const { creator, created } = createFakeCreator();
    const socket = new Socket(creator);

    socket.setSessionActive(true);
    socket.start();
    socket.setSessionActive(false);

    expect(created[0].closed).to.equal(true);
    expect(socket.isRunning()).to.equal(false);
  });
});
