<template>
  <div class="pb-6" style="margin-top: -10px;" :data-testid="testId">
    <v-dialog
      v-model="editDialog"
      :scrim="false"
      content-class="NestedDialog"
      width="400"
    >
      <v-card
        :color="$vuetify.theme.current.dark ? '#212121' : 'white'"
        :data-testid="`${testId}Dialog`"
      >
        <v-card-title :data-testid="`${testId}Dialog-title`"></v-card-title>
        <v-card-text class="pb-0">
          <v-form
            ref="form"
            v-if="editedVar != null"
            :data-testid="`${testId}Dialog-form`"
            @submit.prevent="saveVar()"
          >
            <v-alert
              :model-value="formError != null"
              color="error"
              :data-testid="`${testId}Dialog-error`"
            >{{ formError }}
            </v-alert>

            <v-text-field
              :label="argTitle || $t('arg')"
              v-model.trim="editedVar.name"
              :data-testid="`${testId}Dialog-name`"
              :rules="[(v) => !!v || $t('arg_required')]"
              required
              variant="outlined"
              density="compact"
              autofocus
              @keydown.enter.prevent="saveVar()"
            />

            <div class="text-right mt-2">

              <v-btn
                color="primary"
                v-if="editedVar.type === 'enum'"
                :data-testid="`${testId}Dialog-addValue`"
                @click="addEditedVarValue()"
              >Add Value</v-btn>
            </div>
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="blue-darken-1"
            variant="text"
            @click="editDialog = false"
            :data-testid="`${testId}Dialog-cancel`"
          >
            {{ $t('cancel') }}
          </v-btn>
          <v-btn
            color="blue-darken-1"
            variant="text"
            @click="saveVar()"
            :data-testid="`${testId}Dialog-save`"
          >
            {{ editedVarIndex == null ? $t('add') : $t('save') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <!-- margin: 0 — the reset of Vuetify 2/3 zeroed every margin, the one of Vuetify 4 does not:
         without it the browser's 2px side margin of a fieldset shifts the frame off the fields. -->
    <fieldset style="margin: 0;
                     padding: 0 10px 2px 10px;
                        border-width: 1px;
                        border-color: rgba(133, 133, 133, 0.4);
                        background-color: rgba(133, 133, 133, 0.1);
                     border-radius: 8px;
                     font-size: 12px;"
    >
      <legend
        style="padding: 0 3px;"
        :data-testid="`${testId}-title`"
      >{{ title || $t('Args') }}</legend>
      <v-chip-group column style="margin-top: -4px;">
        <draggable
          v-model="modifiedVars"
          @end="onDragEnd"
          :animation="200"
          class="d-flex flex-wrap"
          ghost-class="chip-ghost"
        >
          <!-- A closable v-chip of Vuetify 3 hides itself on close unless its model is controlled
               (a value and an update handler): the chips are keyed by index, so the hidden one
               would be reused for the next arg. The list (deleteVar) decides what is shown. -->
          <v-chip
            v-for="(v, i) in modifiedVars"
            closable
            :model-value="true"
            @update:model-value="() => {}"
            @click:close="deleteVar(i)"
            :key="i"
            @click="editVar(i)"
            class="draggable-chip"
            :data-testid="`${testId}-chip`"
          >
            <div
              style="
                max-width: 200px;
                overflow: hidden;
                text-overflow: ellipsis;
              "
            >{{ v.name }}</div>
          </v-chip>
        </draggable>
        <v-chip @click="editVar(null)" :data-testid="`${testId}-add`">
          + <span
                class="ml-1"
                v-if="modifiedVars.length === 0"
            >{{ addArgTitle || $t('addArg') }}</span>
        </v-chip>
      </v-chip-group>
    </fieldset>
  </div>
</template>
<style lang="scss">
.draggable-chip {
  cursor: grab;

  &:active {
    cursor: grabbing;
  }
}

.chip-ghost {
  opacity: 0.5;
}
</style>
<script>
import { VueDraggable as draggable } from 'vue-draggable-plus';

export default {
  components: {
    draggable,
  },
  props: {
    vars: Array,
    title: String,
    addArgTitle: String,
    argTitle: String,
    // Builds data-testid of the picker and its parts only, no behaviour.
    testId: { type: String, default: 'argsPicker' },
  },
  emits: ['change'],
  watch: {
    vars(val) {
      this.var = val || [];
      this.fillModifiedVars();
    },
  },

  created() {
    this.fillModifiedVars();
  },

  data() {
    return {
      editDialog: null,
      editedVar: null,
      editedValues: [],
      editedVarIndex: null,
      modifiedVars: null,
      formError: null,
    };
  },
  methods: {
    fillModifiedVars() {
      this.modifiedVars = (this.vars || []).map((v) => ({ name: v }));
    },

    addEditedVarValue() {
      this.editedValues.push({
        name: '',
        value: '',
      });
    },

    editVar(index) {
      this.editedVar = index != null ? { ...this.modifiedVars[index] } : {};

      this.editedValues = [];
      this.editedValues.push(...(this.editedVar.values || []));
      this.editedVar.values = this.editedValues;

      this.editedVarIndex = index;

      if (this.$refs.form) {
        this.$refs.form.resetValidation();
      }

      this.editDialog = true;
    },

    async saveVar() {
      this.formError = null;

      const { valid } = await this.$refs.form.validate();
      if (!valid) {
        return;
      }

      this.editedVar.values = [];

      if (this.editedVarIndex != null) {
        this.modifiedVars[this.editedVarIndex] = this.editedVar;
      } else {
        this.modifiedVars.push(this.editedVar);
      }

      this.editDialog = false;
      this.editedVar = null;
      this.$emit('change', this.modifiedVars.map((x) => x.name));
    },

    deleteVar(index) {
      this.modifiedVars.splice(index, 1);
      this.$emit('change', this.modifiedVars.map((x) => x.name));
    },

    onDragEnd() {
      this.$emit('change', this.modifiedVars.map((x) => x.name));
    },
  },
};
</script>
