<template>
  <div class="pb-6" style="margin-top: -10px;" data-testid="surveyVars">
    <v-dialog
      v-model="editDialog"
      :scrim="false"
      content-class="NestedDialog"
      width="440"
    >
      <v-card
        :color="$vuetify.theme.current.dark ? '#212121' : 'white'"
        @keydown.enter.prevent
        data-testid="surveyVarDialog"
      >
        <v-card-title data-testid="surveyVarDialog-heading"></v-card-title>
        <v-card-text class="pb-0">
          <v-form
            ref="form"
            lazy-validation
            v-if="editedVar != null"
            @submit.prevent="saveVar()"
            data-testid="surveyVarDialog-form"
            :data-state="formError != null ? 'error' : 'ready'"
          >
            <v-alert
              :model-value="formError != null"
              color="error"
              data-testid="surveyVarDialog-error"
            >{{ formError }}
            </v-alert>

            <v-text-field
              :label="$t('name2')"
              v-model.trim="editedVar.name"
              :rules="[(v) => !!v || $t('name_required')]"
              required
              density="compact"
              variant="outlined"
              data-testid="surveyVarDialog-name"
            />

            <v-text-field
              :label="$t('title')"
              v-model="editedVar.title"
              :rules="[(v) => !!v || $t('title_required')]"
              required
              density="compact"
              variant="outlined"
              data-testid="surveyVarDialog-title"
            />

            <v-text-field
              :label="$t('description')"
              v-model="editedVar.description"
              required
              density="compact"
              variant="outlined"
              data-testid="surveyVarDialog-description"
            />

            <v-row>
              <v-col cols="5">
                <v-select
                  v-model="editedVar.type"
                  :label="$t('type')"
                  :items="availableVarTypes"
                  item-value="id"
                  item-title="name"
                  density="compact"
                  variant="outlined"
                  data-testid="surveyVarDialog-type"
                ></v-select>
              </v-col>
              <v-col cols="7">
                <v-select
                  v-model="editedVar.target"
                  :label="$t('survey_var_target')"
                  :items="varTargets"
                  item-value="id"
                  item-title="name"
                  density="compact"
                  variant="outlined"
                  data-testid="surveyVarDialog-target"
                ></v-select>
              </v-col>
            </v-row>

            <v-card
              v-if="editedVar.type === 'enum' || editedVar.type === 'select'"
              style="background: var(--highlighted-card-bg-color)"
              class="mb-4 pt-3"
              data-testid="surveyVarDialog-values"
            >
              <div
                style="
                  position: absolute;
                  background: var(--highlighted-card-bg-color);
                  width: 28px;
                  height: 28px;
                  transform: rotate(45deg);
                  left: 60px;
                  top: -14px;
                  border-radius: 0;
                "
              ></div>

              <v-card-text class="pt-2">
                <v-data-table
                  :items="editedValues"
                  :items-per-page="-1"
                  class="elevation-1 FieldTable"
                  hide-default-footer
                  :no-data-text="$t('noValues')"
                  data-testid="surveyVarDialog-valuesTable"
                  :data-state="editedValues.length === 0 ? 'empty' : 'ready'"
                >
                  <template v-slot:item="props">
                    <tr data-testid="surveyVarDialog-valueRow">
                      <td class="pa-1">
                        <v-text-field
                          variant="solo-inverted"
                          flat
                          hide-details
                          v-model="props.item.name"
                          :label="$t('matchKey')"
                          class="v-text-field--solo--no-min-height"
                          data-testid="surveyVarDialog-valueName"
                        ></v-text-field>
                      </td>
                      <td class="pa-1">
                        <v-text-field
                          variant="solo-inverted"
                          flat
                          hide-details
                          v-model="props.item.value"
                          :label="$t('matchValue')"
                          class="v-text-field--solo--no-min-height"
                          data-testid="surveyVarDialog-valueValue"
                        ></v-text-field>
                      </td>
                      <td style="width: 38px">
                        <v-icon
                          size="small"
                          class="pa-1"
                          @click="removeEditedVarValue(props.item)"
                          data-testid="surveyVarDialog-valueDelete"
                        >
                          mdi-delete
                        </v-icon>
                      </td>
                    </tr>
                  </template>
                </v-data-table>

                <div class="text-right mt-2 mb-2">
                  <v-btn
                    color="primary"
                    @click="addEditedVarValue()"
                    data-testid="surveyVarDialog-addValue"
                  >Add Value</v-btn>
                </div>

                <v-select
                  v-model="editedVar.default_value"
                  :label="$t('default_value')"
                  :items="editedValues"
                  item-value="value"
                  item-title="name"
                  clearable
                  density="compact"
                  variant="outlined"
                  hide-details
                  :multiple="editedVar.type === 'select'"
                  :chips="editedVar.type === 'select'"
                  data-testid="surveyVarDialog-defaultValue"
                >
                  <template
                    v-slot:chip="{ item, index }"
                    v-if="editedVar.type === 'select'">
                    <v-chip
                      size="small"
                      closable
                      data-testid="surveyVarDialog-defaultChip"
                      @click:close="removeDefaultItem(index)"
                      >{{ selectedItemLabel(item) }}
                    </v-chip>
                  </template>
                </v-select>
              </v-card-text>
            </v-card>

            <v-text-field
              type="number"
              v-else-if="editedVar.type === 'int'"
              :label="$t('default_value')"
              v-model="editedVar.default_value"
              density="compact"
              variant="outlined"
              data-testid="surveyVarDialog-defaultValue"
            />

            <v-textarea
              v-else-if="editedVar.type === 'text'"
              :label="$t('default_value')"
              v-model="editedVar.default_value"
              rows="3"
              auto-grow
              density="compact"
              variant="outlined"
              data-testid="surveyVarDialog-defaultValue"
            />

            <v-text-field
              v-else-if="editedVar.type !== 'secret'"
              :label="$t('default_value')"
              v-model="editedVar.default_value"
              density="compact"
              variant="outlined"
              data-testid="surveyVarDialog-defaultValue"
            />
          </v-form>
        </v-card-text>
        <v-card-actions class="py-0">
          <v-checkbox
            :label="$t('required')"
            v-model="editedVar.required"
            v-if="editedVar != null"
            class="ml-1"
            data-testid="surveyVarDialog-required"
          />
          <v-spacer></v-spacer>
          <v-btn
            color="blue-darken-1"
            variant="text"
            @click="editDialog = false"
            data-testid="surveyVarDialog-cancel"
          >
            {{ $t('cancel') }}
          </v-btn>
          <v-btn
            color="blue-darken-1"
            variant="text"
            @click="saveVar()"
            data-testid="surveyVarDialog-save"
          >
            {{ editedVarIndex == null ? $t('add') : $t('save') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <fieldset
      style="
        padding: 0 10px 2px 10px;
        border-width: 1px;
        border-color: rgba(133, 133, 133, 0.4);
        background-color: rgba(133, 133, 133, 0.1);
        border-radius: 8px;
        font-size: 12px;
      "
    >
      <legend
        style="padding: 0 3px"
        data-testid="surveyVars-title"
      >{{ $t('surveyVariables') }}</legend>
      <v-chip-group column style="margin-top: -4px">
        <draggable
          v-model="modifiedVars"
          @end="onDragEnd"
          :animation="200"
          class="d-flex flex-wrap"
          ghost-class="chip-ghost"
        >
          <v-chip
            v-for="(v, i) in modifiedVars"
            closable
            @click:close="deleteVar(i)"
            :key="v.name"
            @click="editVar(i)"
            :color="v.type === 'int' ? '#61e2ff' : 'gray'"
            class="draggable-chip"
            :data-testid="`surveyVars-chip-${v.name}`"
          >
            {{ v.title }}
          </v-chip>
        </draggable>
        <v-chip @click="editVar(null)" data-testid="surveyVars-add">
          + <span class="ml-1" v-if="modifiedVars.length === 0">{{ $t('addVariable') }}</span>
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
    // A workflow survey can not hold a secret: the run stores its answers in
    // clear. The template form keeps the full list.
    allowSecret: { type: Boolean, default: true },
  },
  emits: ['change'],
  computed: {
    availableVarTypes() {
      if (this.allowSecret) return this.varTypes;
      return this.varTypes.filter((t) => t.id !== 'secret');
    },
  },
  watch: {
    vars(val) {
      this.var = val || [];
    },

    'editedVar.type': {
      handler(newType, oldType) {
        if (!this.editedVar || newType === oldType) {
          return;
        }

        if (newType === 'select') {
          this.editedVar.default_value = [];
        } else {
          this.editedVar.default_value = '';
        }

        const isNotListType = newType !== 'enum' && newType !== 'select';
        if (isNotListType) {
          this.editedValues = [];
          this.editedVar.values = this.editedValues;
        }
      },
    },
  },

  created() {
    this.modifiedVars = (this.vars || []).map((v) => ({ ...v }));
  },

  data() {
    return {
      editDialog: null,
      editedVar: null,
      editedValues: [],
      editedVarIndex: null,
      modifiedVars: null,
      varTypes: [
        {
          id: '',
          name: 'String',
        },
        {
          id: 'int',
          name: 'Integer',
        },
        {
          id: 'secret',
          name: 'Secret',
        },
        {
          id: 'enum',
          name: 'Enum',
        },
        {
          id: 'text',
          name: 'Text',
        },
        {
          id: 'select',
          name: 'Select',
        },
      ],
      varTargets: [
        {
          id: '',
          name: 'Extra variable',
        },
        {
          id: 'env',
          name: 'Environment variable',
        },
      ],
      formError: null,
    };
  },
  methods: {
    addEditedVarValue() {
      this.editedValues.push({
        name: '',
        value: '',
      });
    },

    removeEditedVarValue(val) {
      const i = this.editedValues.findIndex((v) => v.name === val.name);
      if (i > -1) {
        this.editedValues.splice(i, 1);
      }
    },

    editVar(index) {
      this.editedVar = index != null ? { ...this.modifiedVars[index] } : {};

      if (!this.editedVar.target) {
        this.editedVar.target = '';
      }

      this.editedValues = [];
      this.editedValues.push(...(this.editedVar.values || []));
      this.editedVar.values = this.editedValues;

      this.editedVar.default_value = this.normalizeDefaultValue(
        this.editedVar.type,
        this.editedVar.default_value,
      );

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

      const typeLabel = this.editedVar.type === 'select' ? 'Select' : 'Enumeration';

      if (this.editedVar.type === 'enum' || this.editedVar.type === 'select') {
        if (this.editedValues.length === 0) {
          this.formError = `${typeLabel} must have values.`;
          return;
        }

        const uniq = new Set(this.editedValues.map((v) => v.name));

        if (this.editedValues.length !== uniq.size) {
          this.formError = `${typeLabel} must have unique names.`;
          return;
        }

        this.editedValues.forEach((v) => {
          if (v.name === '') {
            this.formError = 'Value name cannot be empty.';
          }
        });

        if (this.formError != null) {
          return;
        }
      } else {
        this.editedVar.values = [];
      }

      this.editedVar.default_value = this.normalizeDefaultValue(
        this.editedVar.type,
        this.editedVar.default_value,
      );

      if (this.editedVarIndex != null) {
        this.modifiedVars[this.editedVarIndex] = this.editedVar;
      } else {
        this.modifiedVars.push(this.editedVar);
      }

      this.editDialog = false;
      this.editedVar = null;
      this.$emit('change', this.modifiedVars);
    },

    deleteVar(index) {
      this.modifiedVars.splice(index, 1);
      this.$emit('change', this.modifiedVars);
    },

    // remove one selected default item (for multiple select)
    removeDefaultItem(index) {
      if (!this.editedVar || !Array.isArray(this.editedVar.default_value)) {
        return;
      }

      if (index >= 0 && index < this.editedVar.default_value.length) {
        this.editedVar.default_value.splice(index, 1);
      }
    },

    // label shown in selection chips - item could be the value or an object
    selectedItemLabel(item) {
      if (!item) return '';
      if (typeof item === 'object' && item.name) return item.name;
      const found = this.editedValues.find((v) => v.value === item || v.name === item);
      return (found && found.name) || String(item);
    },

    normalizeDefaultValue(type, value) {
      if (type === 'select') {
        if (Array.isArray(value)) return value;
        return value == null || value === '' ? [] : [value];
      }
      if (Array.isArray(value)) {
        return value.length > 0 ? value[0] : '';
      }
      return value ?? '';
    },

    onDragEnd() {
      this.$emit('change', this.modifiedVars);
    },
  },
};
</script>
