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
          <v-form v-if="editedVar != null">
            <v-text-field
              label="Name"
              v-model="editedVar.name"
              required
            />
            <v-text-field
              label="Title"
              v-model="editedVar.title"
              required
            />
            <v-text-field
              label="Description (Optional)"
              v-model="editedVar.description"
              required
            />
            <v-select
              v-model="editedVar.type"
              label="Type (Optional)"
              :items="varTypes"
              item-value="id"
              item-text="name"
            ></v-select>
            <v-checkbox
              label="Required"
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
            Cancel
          </v-btn>
          <v-btn
            color="blue darken-1"
            text
            @click="saveVar()"
          >
            {{ editedVarIndex == null ? 'Add' : 'Save' }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <fieldset style="padding: 0 10px 5px 10px;
                     border: 1px solid rgba(0, 0, 0, 0.38);
                     border-radius: 4px;
                     font-size: 12px;"
              :style="{
                       'border-color': $vuetify.theme.dark ?
                         'rgba(200, 200, 200, 0.38)' :
                         'rgba(0, 0, 0, 0.38)'
                     }">
      <legend style="padding: 0 3px;">Survey Variables</legend>
      <v-chip-group column>
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
        <v-chip @click="editVar(null)">+</v-chip>
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
      editedVarIndex: null,
      modifiedVars: null,
      varTypes: [{
        id: '',
        name: 'String',
      }, {
        id: 'int',
        name: 'Integer',
      }],
    };
  },
  methods: {
    editVar(index) {
      this.editedVar = index != null ? { ...this.modifiedVars[index] } : {};
      this.editedVarIndex = index;
      this.editDialog = true;
    },

    saveVar() {
      if (this.editedVarIndex != null) {
        this.modifiedVars[this.editedVarIndex] = this.editedVar;
      } else {
        this.modifiedVars.push(this.editedVar);
      }
      this.editDialog = false;
      this.editVarIndex = null;
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
