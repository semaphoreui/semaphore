<template>
  <v-card
      style="background: var(--highlighted-card-bg-color)"
      class="mb-6 pt-3"
  >
    <div
        style="
              position: absolute;
              background: var(--highlighted-card-bg-color);
              width: 28px;
              height: 28px;
              transform: rotate(45deg);
              left: calc(50% - 14px);
              top: -14px;
              border-radius: 0;
            "
    ></div>
    <v-card-text class="pb-0">
      <slot></slot>
    </v-card-text>
  </v-card>
</template>
