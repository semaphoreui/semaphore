<template>
  <div class="pb-6">
    <v-dialog
      v-model="editDialog"
      hide-overlay
      width="300"
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
              v-if="editedVar.type === 'enum'"
              :items="editedValues"
              :items-per-page="-1"
              class="elevation-1"
              hide-default-footer
              no-data-text="No values"
            >
              <template v-slot:item="props">
                <tr>
                  <td class="pa-1">
                    <v-text-field
                      solo-inverted
                      flat
                      hide-details
                      v-model="props.item.name"
                      class="v-text-field--solo--no-min-height"
                    ></v-text-field>
                  </td>
                  <td class="pa-1">
                    <v-text-field
                      solo-inverted
                      flat
                      hide-details
                      v-model="props.item.value"
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
                v-if="editedVar.type === 'enum'"
                @click="addEditedVarValue()"
              >Add Value</v-btn>
            </div>

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
                     border: 1px solid rgba(0, 0, 0, 0.38);
                     border-radius: 4px;
                     font-size: 12px;"
              :style="{
                       'border-color': $vuetify.theme.dark ?
                         'rgba(200, 200, 200, 0.38)' :
                         'rgba(0, 0, 0, 0.38)'
                     }">
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
<style lang="scss">

</style>
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

      if (this.editedVar.type === 'enum') {
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
  },
};
</script>
