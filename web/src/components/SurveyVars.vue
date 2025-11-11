<template>
  <div class="pb-6" style="margin-top: -10px;">
    <v-dialog
      v-model="editDialog"
      hide-overlay
      width="400"
    >
      <v-card :color="$vuetify.theme.dark ? '#212121' : 'white'">
        <v-card-title></v-card-title>
        <v-card-text class="pb-0">
          <v-form
            ref="form"
            lazy-validation
            v-if="editedVar != null"
          >
            <v-alert
              :value="formError"
              color="error"
            >{{ formError }}
            </v-alert>

            <v-text-field
              :label="$t('name2')"
              v-model.trim="editedVar.name"
              :rules="[(v) => !!v || $t('name_required')]"
              required
            />

            <v-text-field
              :label="$t('title')"
              v-model="editedVar.title"
              :rules="[(v) => !!v || $t('title_required')]"
              required
            />

            <v-text-field
              :label="$t('description')"
              v-model="editedVar.description"
              required
            />

            <v-select
              v-model="editedVar.type"
              :label="$t('type')"
              :items="varTypes"
              item-value="id"
              item-text="name"
            ></v-select>

            <v-data-table
              v-if="editedVar.type === 'enum' || editedVar.type === 'select'"
              :items="editedValues"
              :items-per-page="-1"
              class="elevation-1 FieldTable"
              hide-default-footer
              :no-data-text="$t('noValues')"
            >
              <template v-slot:item="props">
                <tr>
                  <td class="pa-1">
                    <v-text-field
                      solo-inverted
                      flat
                      hide-details
                      v-model="props.item.name"
                      :label="$t('matchKey')"
                      class="v-text-field--solo--no-min-height"
                    ></v-text-field>
                  </td>
                  <td class="pa-1">
                    <v-text-field
                      solo-inverted
                      flat
                      hide-details
                      v-model="props.item.value"
                      :label="$t('matchValue')"
                      class="v-text-field--solo--no-min-height"
                    ></v-text-field>
                  </td>
                  <td style="width: 38px;">
                    <v-icon
                      small
                      class="pa-1"
                      @click="removeEditedVarValue(props.item)"
                    >
                      mdi-delete
                    </v-icon>
                  </td>
                </tr>
              </template>
            </v-data-table>

            <div class="text-right mt-2">
              <v-btn
                color="primary"
                class="d-block"
                v-if="editedVar.type === 'enum' || editedVar.type === 'select'"
                @click="addEditedVarValue()"
              >Add Value</v-btn>
            </div>

            <v-select
              v-if="editedVar.type === 'enum' || editedVar.type === 'select'"
              v-model="editedVar.default_value"
              :label="$t('default_value')"
              :items="editedValues"
              item-value="value"
              item-text="name"
              clearable
              :multiple="editedVar.type === 'select'"
              :chips="editedVar.type === 'select'"
            >
              <template
                v-slot:selection="{ item, index }"
                v-if="editedVar.type === 'select'">
                <v-chip
                  small
                  close
                  @click:close="removeDefaultItem(index)"
                  >{{ selectedItemLabel(item) }}
                </v-chip>
              </template>
            </v-select>

            <v-text-field
              type="number"
              v-else-if="editedVar.type === 'int'"
              :label="$t('default_value')"
              v-model="editedVar.default_value"
            />

            <v-text-field
              v-else-if="editedVar.type !== 'secret'"
              :label="$t('default_value')"
              v-model="editedVar.default_value"
            />

            <v-checkbox
              :label="$t('required')"
              v-model="editedVar.required"
            />
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="blue darken-1"
            text
            @click="editDialog = false"
          >
            {{ $t('cancel') }}
          </v-btn>
          <v-btn
            color="blue darken-1"
            text
            @click="saveVar()"
          >
            {{ editedVarIndex == null ? $t('add') : $t('save') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <fieldset style="padding: 0 10px 2px 10px;
                        border-width: 1px;
                        border-color: rgba(133, 133, 133, 0.4);
                        background-color: rgba(133, 133, 133, 0.1);
                     border-radius: 8px;
                     font-size: 12px;"
    >
      <legend style="padding: 0 3px;">{{ $t('surveyVariables') }}</legend>
      <v-chip-group column style="margin-top: -4px;">
        <v-chip
          v-for="(v, i) in modifiedVars"
          close
          @click:close="deleteVar(i)"
          :key="v.name"
          @click="editVar(i)"
          :color="v.type === 'int' ? '#61e2ff' : 'gray'"
        >
          {{ v.title }}
        </v-chip>
        <v-chip @click="editVar(null)">
          + <span class="ml-1" v-if="modifiedVars.length === 0">{{ $t('addVariable') }}</span>
        </v-chip>
      </v-chip-group>
    </fieldset>
  </div>
</template>
<script>
export default {
  props: {
    vars: Array,
  },
  watch: {
    vars(val) {
      this.var = val || [];
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
      varTypes: [{
        id: '',
        name: 'String',
      }, {
        id: 'int',
        name: 'Integer',
      }, {
        id: 'secret',
        name: 'Secret',
      }, {
        id: 'enum',
        name: 'Enum',
      }, {
        id: 'select',
        name: 'Select',
      }],
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

      this.editedValues = [];
      this.editedValues.push(...(this.editedVar.values || []));
      this.editedVar.values = this.editedValues;

      // normalize default_value for select vs enum
      if (this.editedVar.type === 'select') {
        if (this.editedVar.default_value == null) {
          this.editedVar.default_value = [];
        } else if (!Array.isArray(this.editedVar.default_value)) {
          const dv = this.editedVar.default_value;
          this.editedVar.default_value = dv === '' ? [] : [dv];
        }
      } else if (Array.isArray(this.editedVar.default_value)) {
        this.editedVar.default_value = this.editedVar.default_value.length > 0
          ? this.editedVar.default_value[0]
          : '';
      }

      this.editedVarIndex = index;

      if (this.$refs.form) {
        this.$refs.form.resetValidation();
      }

      this.editDialog = true;
    },

    saveVar() {
      this.formError = null;

      if (!this.$refs.form.validate()) {
        return;
      }

      if (this.editedVar.type === 'enum' || this.editedVar.type === 'select') {
        if (this.editedValues.length === 0) {
          this.formError = 'Enumeration must have values.';
          return;
        }

        const uniq = new Set(this.editedValues.map((v) => v.name));

        if (this.editedValues.length !== uniq.size) {
          this.formError = 'Enumeration values must have unique names.';
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

      // normalize default_value before saving: select keeps array, others use string
      if (this.editedVar.type === 'select') {
        if (!Array.isArray(this.editedVar.default_value)) {
          const dv = this.editedVar.default_value;
          this.editedVar.default_value = dv == null || dv === '' ? [] : [dv];
        }
      } else if (Array.isArray(this.editedVar.default_value)) {
        this.editedVar.default_value = this.editedVar.default_value.length > 0
          ? this.editedVar.default_value[0]
          : '';
      }

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

    onTypeChange(newType) {
      if (!this.editedVar) return;
      if (newType === 'select') {
        if (!Array.isArray(this.editedVar.default_value)) {
          const dv = this.editedVar.default_value;
          this.editedVar.default_value = dv == null || dv === '' ? [] : [dv];
        }
      } else if (Array.isArray(this.editedVar.default_value)) {
        this.editedVar.default_value = this.editedVar.default_value.length > 0
          ? this.editedVar.default_value[0]
          : '';
      }
    },
  },
};
</script>
