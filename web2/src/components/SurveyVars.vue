<template>
  <div class="pb-6">
    <v-dialog
      v-model="editDialog"
      hide-overlay
      width="300"
    >
      <v-card>
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
                     font-size: 12px;">
      <legend style="padding: 0 3px;">Survey Variables</legend>
      <v-chip-group column>
        <v-chip
          v-for="(v, i) in vars"
          close
          @click:close="deleteVar(i)"
          :key="v.name"
          @click="editVar(i)"
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
    json: String,
  },
  watch: {
    json(val) {
      this.var = JSON.parse(val || '[]');
    },
  },
  created() {
    this.vars = JSON.parse(this.json || '[]');
  },
  data() {
    return {
      editDialog: null,
      editedVar: null,
      editedVarIndex: null,
      vars: null,
    };
  },
  methods: {
    editVar(index) {
      this.editedVar = index != null ? { ...this.vars[index] } : {};
      this.editedVarIndex = index;
      this.editDialog = true;
    },

    saveVar() {
      if (this.editedVarIndex != null) {
        this.vars[this.editedVarIndex] = this.editedVar;
      } else {
        this.vars.push(this.editedVar);
      }
      this.editDialog = false;
      this.editVarIndex = null;
      this.editedVar = null;
      this.$emit('change', JSON.stringify(this.vars));
    },

    deleteVar(index) {
      this.vars.splice(index, 1);
      this.$emit('change', JSON.stringify(this.vars));
    },
  },
};
</script>
