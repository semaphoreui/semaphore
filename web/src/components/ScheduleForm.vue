<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="templates && item != null"
  >
    <v-alert
      v-model="showInfo"
      color="info"
      text
      class="mb-6"
    >
      Use environment variable <code>SEMAPHORE_SCHEDULE_TIMEZONE</code> or config param
      <code>schedule.timezone</code> to set timezone for Schedule.
    </v-alert>

    <v-alert
      :value="formError"
      color="error"
      class="pb-2"
    >{{ formError }}
    </v-alert>

    <v-text-field
      v-model="item.name"
      :label="$t('Name')"
      :rules="[v => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    ></v-text-field>

    <v-autocomplete
      v-model="item.template_id"
      :label="$t('Template')"
      :items="templates"
      item-value="id"
      :item-text="(itm) => itm.name"
      :rules="[v => !!v || $t('template_required')]"
      required
      :disabled="formSaving"
      outlined
      dense
    />

    <v-card
      style="background: rgba(133, 133, 133, 0.06)"
      v-if="item.template_id"
      class="mb-8"
    >
      <v-card-text>
        <TaskParamsForm
          :template="templates.find(t => t.id === item.template_id)"
          v-model="item.task_params"
        />

      </v-card-text>
    </v-card>

    <v-switch
      v-model="rawCron"
      label="Show cron format"
      :disabled="disableRawCron"
    />

    <v-text-field
      v-if="rawCron"
      v-model="item.cron_format"
      :label="$t('Cron')"
      :rules="[v => !!v || $t('Cron required')]"
      required
      :disabled="formSaving"
      @input="refreshCheckboxes()"
      :suffix="timezone + ' time'"
      outlined
      :error="cronFormatError != null"
      :error-messages="cronFormatError"
      dense
    ></v-text-field>

    <div v-else>
      <v-select
        v-model="timing"
        :label="$t('Timing')"
        :items="TIMINGS"
        item-value="id"
        item-text="title"
        :rules="[v => !!v || $t('template_required')]"
        required
        :disabled="formSaving"
        @change="refreshCron()"
        outlined
        hide-details
        dense
      />

      <div v-if="['yearly'].includes(timing)">
        <div class="mt-4">Months</div>
        <div class="d-flex flex-wrap">
          <v-checkbox
            class="mr-2 mt-0 ScheduleCheckbox"
            v-for="m in MONTHS"
            :key="m.id"
            :value="m.id"
            :label="m.title"
            v-model="months"
            color="white"
            :class="{'ScheduleCheckbox--active': months.includes(m.id)}"
            @change="refreshCron()"
          ></v-checkbox>
        </div>
      </div>

      <div v-if="['weekly'].includes(timing)">
        <div class="mt-4">Weekdays</div>
        <div class="d-flex flex-wrap">
          <v-checkbox
            class="mr-2 mt-0 ScheduleCheckbox"
            v-for="d in WEEKDAYS" :key="d.id"
            :value="d.id"
            :label="d.title"
            v-model="weekdays"
            color="white"
            :class="{'ScheduleCheckbox--active': weekdays.includes(d.id)}"
            @change="refreshCron()"
          ></v-checkbox>
        </div>
      </div>

      <div v-if="['yearly', 'monthly'].includes(timing)">
        <div class="mt-4">Days</div>
        <div class="d-flex flex-wrap">
          <v-checkbox
            class="mr-2 mt-0 ScheduleCheckbox"
            v-for="d in 31"
            :key="d"
            :value="d"
            :label="`${d}`"
            v-model="days"
            color="white"
            :class="{'ScheduleCheckbox--active': days.includes(d)}"
            @change="refreshCron()"
          ></v-checkbox>
        </div>
      </div>

      <div v-if="['yearly', 'monthly', 'weekly', 'daily'].includes(timing)">
        <div class="mt-4 d-flex justify-space-between">
          <span>Hours</span>
          <b style="color: red;">{{ timezone + ' time' }}</b>
        </div>
        <div class="d-flex flex-wrap">
          <v-checkbox
            class="mr-2 mt-0 ScheduleCheckbox"
            v-for="h in 24"
            :key="h - 1"
            :value="h - 1"
            :label="`${h - 1}`"
            v-model="hours"
            color="white"
            :class="{'ScheduleCheckbox--active': hours.includes(h - 1)}"
            @change="refreshCron()"
          ></v-checkbox>
        </div>
      </div>

      <div>
        <div class="mt-4">Minutes</div>
        <div class="d-flex flex-wrap">
          <v-checkbox
            class="mr-2 mt-0 ScheduleCheckbox"
            v-for="m in MINUTES"
            :key="m.id"
            :value="m.id"
            :label="m.title"
            v-model="minutes"
            color="white"
            :class="{'ScheduleCheckbox--active': minutes.includes(m.id)}"
            @change="refreshCron()"
          ></v-checkbox>
        </div>
      </div>
    </div>

    <div
      class="text-center text-subtitle-1 mb-3"
      :class="{'mt-8': !rawCron, 'mt-3': rawCron}"
      style="color: limegreen; font-weight: bold;"
    >
      Next run time
    </div>

    <v-simple-table class="TaskDetails__table text-sub mb-2">
      <template v-slot:default>
        <thead>
        <tr>
          <th>Time Zone</th>
          <th>Date</th>
          <th>Time</th>
        </tr>
        </thead>
        <tbody>
        <tr>
          <td>{{ timezone }}</td>
          <td>{{ nextRunUtcDate }}</td>
          <td>{{ nextRunUtcTime }}</td>
        </tr>
        <tr>
          <td>{{ localTimezone }}</td>
          <td>{{ nextRunLocalDate }}</td>
          <td>{{ nextRunLocalTime }}</td>
        </tr>
        </tbody>
      </template>
    </v-simple-table>

    <v-checkbox
      style="position: absolute; bottom: 15px; left: 22px;"
      v-model="item.active"
      hide-details
    >
      <template v-slot:label>
        {{ $t('enabled') }}
      </template>
    </v-checkbox>

  </v-form>
</template>

<style lang="scss">
.ScheduleCheckbox {

  .v-input__slot {
    padding: 4px 6px;
    font-weight: bold;
    border-radius: 6px;
  }

  .v-messages {
    display: none;
  }

  &.theme--light {
    .v-input__slot {
      background: #e4e4e4;
    }
  }

  &.theme--dark {
    .v-input__slot {
      background: gray;
    }
  }
}

.ScheduleCheckbox--active {
  .v-input__slot {
    background: #4caf50 !important;
  }

  .v-label {
    color: white;
  }
}

</style>

<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';

import { CronExpression, CronExpressionParser, CronFieldCollection } from 'cron-parser';
import { getErrorMessage } from '@/lib/error';
import TaskParamsForm from '@/components/TaskParamsForm.vue';

const MONTHS = [{
  id: 1,
  title: 'Jan',
}, {
  id: 2,
  title: 'Feb',
}, {
  id: 3,
  title: 'March',
}, {
  id: 4,
  title: 'April',
}, {
  id: 5,
  title: 'May',
}, {
  id: 6,
  title: 'June',
}, {
  id: 7,
  title: 'July',
}, {
  id: 8,
  title: 'August',
}, {
  id: 9,
  title: 'September',
}, {
  id: 10,
  title: 'October',
}, {
  id: 11,
  title: 'November',
}, {
  id: 12,
  title: 'December',
}];

const TIMINGS = [{
  id: 'yearly',
  title: 'Yearly',
}, {
  id: 'monthly',
  title: 'Monthly',
}, {
  id: 'weekly',
  title: 'Weekly',
}, {
  id: 'daily',
  title: 'Daily',
}, {
  id: 'hourly',
  title: 'Hourly',
}];

const WEEKDAYS = [{
  id: 0,
  title: 'Sunday',
}, {
  id: 1,
  title: 'Monday',
}, {
  id: 2,
  title: 'Tuesday',
}, {
  id: 3,
  title: 'Wednesday',
}, {
  id: 4,
  title: 'Thursday',
}, {
  id: 5,
  title: 'Friday',
}, {
  id: 6,
  title: 'Saturday',
}];

const MINUTES = [
  { id: 0, title: ':00' },
  { id: 5, title: ':05' },
  { id: 10, title: ':10' },
  { id: 15, title: ':15' },
  { id: 20, title: ':20' },
  { id: 25, title: ':25' },
  { id: 30, title: ':30' },
  { id: 35, title: ':35' },
  { id: 40, title: ':40' },
  { id: 45, title: ':45' },
  { id: 50, title: ':50' },
  { id: 55, title: ':55' },
];

function formatDateInTZ(date, tz) {
  if (date == null) {
    return '—';
  }
  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone: tz,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).formatToParts(date);

  const get = (type) => parts.find((p) => p.type === type)?.value;

  return `${get('year')}-${get('month')}-${get('day')}`;
}

function formatTimeInTZ(date, tz) {
  if (date == null) {
    return '—';
  }

  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone: tz,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).formatToParts(date);

  const get = (type) => parts.find((p) => p.type === type)?.value;

  return `${get('hour')}:${get('minute')}`;
}

export default {
  components: { TaskParamsForm },
  mixins: [ItemFormBase],

  data() {
    return {
      templates: null,
      timing: 'hourly',
      TIMINGS,
      MONTHS,
      WEEKDAYS,
      MINUTES,
      minutes: [],
      hours: [],
      days: [],
      months: [],
      weekdays: [],
      rawCron: false,
      disableRawCron: false,
      showInfo: true,
      cronFormatError: null,
    };
  },

  watch: {
    rawCron(val) {
      if (val) {
        localStorage.removeItem('schedule__raw_cron');
      } else {
        localStorage.setItem('schedule__raw_cron', '1');
      }
    },

    showInfo(val) {
      if (val) {
        localStorage.removeItem('schedule__hide_info');
      } else {
        localStorage.setItem('schedule__hide_info', '1');
      }
    },
  },

  async created() {
    this.showInfo = localStorage.getItem('schedule_hide_info') !== '1';
    this.rawCron = localStorage.getItem('schedule__raw_cron') !== '1';

    this.templates = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/templates`,
      responseType: 'json',
    })).data;
  },

  props: {
    timezone: String,
  },

  computed: {
    localTimezone() {
      return 'Local';
    },

    nextRunUtcDate() {
      return formatDateInTZ(this.nextRunTime(), this.timezone);
    },

    nextRunLocalDate() {
      const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
      return formatDateInTZ(this.nextRunTime(), tz);
    },

    nextRunUtcTime() {
      return formatTimeInTZ(this.nextRunTime(), this.timezone);
    },

    nextRunLocalTime() {
      const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
      return formatTimeInTZ(this.nextRunTime(), tz);
    },
  },

  methods: {
    nextRunTime() {
      try {
        return CronExpressionParser.parse(this.item.cron_format, {
          tz: this.timezone,
        }).next().toDate();
      } catch {
        return null;
      }
    },

    refreshCheckboxes() {
      // if (!/test/.test(this.item.cron_format)) {
      //   this.rawCron = true;
      //   this.disableRawCron = true;
      // } else {
      //   this.disableRawCron = false;
      // }

      this.cronFormatError = null;
      this.disableRawCron = false;

      let cron;
      try {
        cron = CronExpressionParser.parse(this.item.cron_format, {
          tz: this.timezone,
        });
      } catch (err) {
        this.cronFormatError = getErrorMessage(err);
        this.rawCron = true;
        this.disableRawCron = true;
        return;
      }

      const fields = cron.fields; // JSON.parse(JSON.stringify(cron.fields));

      this.months = [];
      this.weekdays = [];
      this.hours = [];
      this.minutes = [];

      if (this.isHourly(this.item.cron_format)) {
        this.minutes = fields.minute.values;
        this.timing = 'hourly';
      } else {
        this.minutes = [];
      }

      if (this.isDaily(this.item.cron_format)) {
        this.hours = fields.hour.values;
        this.timing = 'daily';
      } else {
        this.hours = [];
      }

      if (this.isWeekly(this.item.cron_format)) {
        this.weekdays = fields.dayOfWeek.values;
        this.timing = 'weekly';
      } else {
        this.weekdays = [];
      }

      if (this.isMonthly(this.item.cron_format)) {
        this.days = fields.dayOfMonth.values;
        this.timing = 'monthly';
      } else {
        this.months = [];
      }

      if (this.isYearly(this.item.cron_format)) {
        this.months = fields.month.values;
        this.timing = 'yearly';
      }
    },

    afterLoadData() {
      if (this.isNew) {
        this.item.cron_format = '* * * * *';
      }

      this.refreshCheckboxes();
    },

    isWeekly(s) {
      return /^\S+\s\S+\s\S+\s\S+\s[^*]\S*$/.test(s);
    },

    isYearly(s) {
      return /^\S+\s\S+\s\S+\s[^*]\S*\s\S+$/.test(s);
    },

    isMonthly(s) {
      return /^\S+\s\S+\s[^*]\S*\s\S+\s\S+$/.test(s);
    },

    isDaily(s) {
      return /^\S+\s[^*]\S*\s\S+\s\S+\s\S+$/.test(s);
    },

    isHourly(s) {
      return /^[^*]\S*\s\S+\s\S+\s\S+\s\S+$/.test(s);
    },

    refreshCron() {
      const fields = {};

      switch (this.timing) {
        case 'hourly':
          this.months = [];
          this.weekdays = [];
          this.days = [];
          this.hours = [];
          break;
        case 'daily':
          this.days = [];
          this.months = [];
          this.weekdays = [];
          break;
        case 'monthly':
          this.months = [];
          this.weekdays = [];
          break;
        case 'weekly':
          this.months = [];
          this.days = [];
          break;
        default:
          break;
      }

      if (this.months.length > 0) {
        fields.month = this.months;
      }

      if (this.weekdays.length > 0) {
        fields.dayOfWeek = this.weekdays;
      }

      if (this.days.length > 0) {
        fields.dayOfMonth = this.days;
      }

      if (this.hours.length > 0) {
        fields.hour = this.hours;
      }

      if (this.minutes.length > 0) {
        fields.minute = this.minutes;
      }

      const origFields = CronExpressionParser.parse('* * * * *').fields;
      const modFields = CronFieldCollection.from(origFields, fields);
      const exp = CronExpression.fieldsToExpression(modFields);
      this.item.cron_format = exp.stringify();
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/schedules`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/schedules/${this.itemId}`;
    },

  },
};
</script>
