<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null && templates != null"
  >
    <v-alert :value="formError" color="error" class="mb-4">{{ formError }}</v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('name')"
      :rules="[(v) => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    />

    <v-text-field
      v-model="item.description"
      :label="$t('description')"
      :disabled="formSaving"
      outlined
      dense
    />

    <div class="text-subtitle-1 mt-2 mb-2 d-flex align-center">
      <span>{{ $t('workflowNodes') }}</span>
      <v-spacer />
      <v-btn
        small
        color="primary"
        @click="addNode()"
        :disabled="formSaving"
      >
        <v-icon left small>mdi-plus</v-icon>{{ $t('workflowAddNode') }}
      </v-btn>
    </div>

    <v-alert
      v-if="(item.nodes || []).length === 0"
      type="info"
      text
      dense
    >{{ $t('workflowNoNodes') }}</v-alert>

    <v-card
      v-for="(node, idx) in item.nodes"
      :key="`node-${idx}`"
      outlined
      class="mb-3 pa-3"
    >
      <div class="d-flex align-center mb-2">
        <span class="text-caption">
          <strong>{{ $t('workflowNodeId') }}:</strong> {{ node.id }}
        </span>
        <v-spacer />
        <v-btn
          icon
          small
          @click="removeNode(idx)"
          :disabled="formSaving"
        >
          <v-icon small>mdi-delete</v-icon>
        </v-btn>
      </div>
      <v-autocomplete
        v-model="node.template_id"
        :items="templates"
        item-value="id"
        item-text="name"
        :label="$t('taskTemplate')"
        :rules="[(v) => !!v || $t('workflowTemplateRequired')]"
        :disabled="formSaving"
        outlined
        dense
        hide-details="auto"
      />
    </v-card>

    <div class="text-subtitle-1 mt-4 mb-2 d-flex align-center">
      <span>{{ $t('workflowEdges') }}</span>
      <v-spacer />
      <v-btn
        small
        color="primary"
        @click="addEdge()"
        :disabled="formSaving || (item.nodes || []).length < 2"
      >
        <v-icon left small>mdi-plus</v-icon>{{ $t('workflowAddEdge') }}
      </v-btn>
    </div>

    <v-alert
      v-if="(item.edges || []).length === 0"
      type="info"
      text
      dense
    >{{ $t('workflowNoEdges') }}</v-alert>

    <v-card
      v-for="(edge, idx) in item.edges"
      :key="`edge-${idx}`"
      outlined
      class="mb-3 pa-3"
    >
      <div class="d-flex">
        <v-select
          v-model="edge.source_node_id"
          :items="nodeOptions"
          item-value="value"
          item-text="text"
          :label="$t('workflowEdgeSource')"
          :rules="[(v) => !!v || $t('workflowEdgeNodeRequired')]"
          :disabled="formSaving"
          outlined
          dense
          hide-details="auto"
          class="mr-2"
        />
        <v-select
          v-model="edge.destination_node_id"
          :items="nodeOptions"
          item-value="value"
          item-text="text"
          :label="$t('workflowEdgeDestination')"
          :rules="[(v) => !!v || $t('workflowEdgeNodeRequired')]"
          :disabled="formSaving"
          outlined
          dense
          hide-details="auto"
          class="mr-2"
        />
        <v-select
          v-model="edge.condition"
          :items="conditionOptions"
          item-value="value"
          item-text="text"
          :label="$t('workflowEdgeCondition')"
          :rules="[(v) => !!v || $t('workflowEdgeConditionRequired')]"
          :disabled="formSaving"
          outlined
          dense
          hide-details="auto"
          class="mr-2"
        />
        <v-btn
          icon
          small
          @click="removeEdge(idx)"
          :disabled="formSaving"
          class="mt-1"
        >
          <v-icon small>mdi-delete</v-icon>
        </v-btn>
      </div>
    </v-card>
  </v-form>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';

export default {
  mixins: [ItemFormBase],
  data() {
    return {
      templates: null,
    };
  },
  computed: {
    nodeOptions() {
      return (this.item?.nodes || []).map((n, i) => {
        const tpl = this.templates?.find((t) => t.id === n.template_id);
        return {
          value: n.id,
          text: tpl ? `#${n.id} — ${tpl.name}` : `#${n.id}`,
          index: i,
        };
      });
    },
    conditionOptions() {
      return [
        { value: 'on_success', text: this.$t('workflowConditionOnSuccess') },
        { value: 'on_failure', text: this.$t('workflowConditionOnFailure') },
        { value: 'always', text: this.$t('workflowConditionAlways') },
      ];
    },
  },
  async created() {
    this.templates = await this.loadProjectResources('templates');
  },
  methods: {
    getNewItem() {
      return {
        name: '',
        description: '',
        nodes: [],
        edges: [],
      };
    },
    afterLoadData() {
      if (!Array.isArray(this.item.nodes)) this.item.nodes = [];
      if (!Array.isArray(this.item.edges)) this.item.edges = [];
    },
    nextNodeId() {
      const ids = (this.item.nodes || []).map((n) => n.id || 0);
      return (ids.length === 0 ? 0 : Math.max(...ids)) + 1;
    },
    addNode() {
      this.item.nodes.push({
        id: this.nextNodeId(),
        template_id: null,
      });
    },
    removeNode(idx) {
      const removed = this.item.nodes[idx];
      this.item.nodes.splice(idx, 1);
      // Drop edges referencing the removed node
      this.item.edges = (this.item.edges || []).filter(
        (e) => e.source_node_id !== removed.id && e.destination_node_id !== removed.id,
      );
    },
    addEdge() {
      const nodes = this.item.nodes || [];
      this.item.edges.push({
        source_node_id: nodes[0]?.id || null,
        destination_node_id: nodes[1]?.id || null,
        condition: 'on_success',
      });
    },
    removeEdge(idx) {
      this.item.edges.splice(idx, 1);
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/workflows`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/workflows/${this.itemId}`;
    },
  },
};
</script>
