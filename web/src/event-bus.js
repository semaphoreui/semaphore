// Application-wide event bus with the Vue 2 instance event API ($on, $once, $off, $emit),
// independent of Vue: Vue 3 instances have no $on/$off/$once.
//
// Handlers run synchronously in subscription order, with the bus as `this`. As in Vue 2, an
// error thrown by a handler, or the rejection of the promise it returns, does not reach the
// caller of $emit: it is logged with console.error and the remaining handlers still run.

const events = Object.create(null);

function logError(err) {
  // eslint-disable-next-line no-console
  console.error(err);
}

function invoke(handler, context, args) {
  try {
    const res = handler.apply(context, args);
    if (res && typeof res.then === 'function' && typeof res.catch === 'function') {
      res.catch(logError);
    }
  } catch (err) {
    logError(err);
  }
}

const EventBus = {
  $on(event, fn) {
    if (Array.isArray(event)) {
      event.forEach((e) => this.$on(e, fn));
    } else {
      (events[event] || (events[event] = [])).push(fn);
    }
    return this;
  },

  $once(event, fn) {
    const bus = this;
    function on(...args) {
      bus.$off(event, on);
      fn.apply(bus, args);
    }
    on.fn = fn;
    return this.$on(event, on);
  },

  $off(event, fn) {
    if (event === undefined) {
      Object.keys(events).forEach((name) => delete events[name]);
      return this;
    }
    if (Array.isArray(event)) {
      event.forEach((e) => this.$off(e, fn));
      return this;
    }
    const cbs = events[event];
    if (!cbs) {
      return this;
    }
    if (fn === undefined) {
      delete events[event];
      return this;
    }
    for (let i = cbs.length - 1; i >= 0; i -= 1) {
      if (cbs[i] === fn || cbs[i].fn === fn) {
        cbs.splice(i, 1);
        break;
      }
    }
    return this;
  },

  $emit(event, ...args) {
    const cbs = events[event];
    if (cbs) {
      cbs.slice().forEach((cb) => invoke(cb, this, args));
    }
    return this;
  },
};

export default EventBus;
