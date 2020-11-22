import Socket from '@/lib/Socket';

const socket = new Socket(() => {
  const protocol = document.location.protocol === 'https:' ? 'wss' : 'ws';
  return new WebSocket(`${protocol}://${document.location.host}/api/ws`);
});

socket.start();

export default socket;
