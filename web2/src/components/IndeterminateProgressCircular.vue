<template>
  <v-progress-circular
    color="white"
    class="indeterminate-progress-circular mr-2"
    size="20"
    width="10"
    :rotate="rotate"
    :value="value"
  >
  </v-progress-circular>
</template>
<style lang="scss">
  .indeterminate-progress-circular {
    .v-progress-circular__overlay {
      transition: 0s !important;
    }
  }
</style>
<script>
class IndeterminateTimer {
  constructor() {
    this.listeners = {};
    this.direction = 1;
    this.value = 0;
    this.rotate = 0;
  }

  start() {
    const STEP = 1;
    const self = this;

    self.valueTimer = setInterval(() => {
      if (self.direction === 1 && self.value >= 100) {
        self.direction = -1;
      } else if (self.direction === -1 && self.value <= 0) {
        self.direction = 1;
      }
      if (self.direction === 1) {
        self.rotate += STEP;
        self.value += STEP;
      } else {
        self.rotate += STEP * 5;
        self.value += -STEP;
      }

      if (self.rotate > 360) {
        self.rotate %= 360;
      }

      Object.keys(self.listeners).forEach((id) => {
        const listener = self.listeners[id];
        listener({
          value: self.value,
          rotate: self.rotate,
        });
      });
    }, 50);
  }

  stop() {
    clearInterval(this.valueTimer);
  }

  addListener(callback) {
    if (Object.keys(this.listeners).length === 0) {
      this.start();
    }
    const id = Math.floor(Math.random() * 100000000);
    this.listeners[id] = callback;
    return id;
  }

  removeListener(id) {
    delete this.listeners[id];
    if (Object.keys(this.listeners).length === 0) {
      this.stop();
    }
  }
}

const indeterminateTimer = new IndeterminateTimer();

export default {
  data() {
    return {
      value: null,
      rotate: null,
      listenerId: null,
    };
  },

  mounted() {
    this.value = indeterminateTimer.value;
    this.rotate = indeterminateTimer.rotate;
    this.listenerId = indeterminateTimer.addListener(({ value, rotate }) => {
      this.value = value;
      this.rotate = rotate;
    });
  },

  beforeDestroy() {
    indeterminateTimer.removeListener(this.listenerId);
  },
};
</script>
