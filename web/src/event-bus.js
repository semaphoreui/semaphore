import mitt from 'mitt';

const emitter = mitt();

export default {
  $on: emitter.on.bind(emitter),
  $off: emitter.off.bind(emitter),
  $emit: emitter.emit.bind(emitter),
};
