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
      dismissible
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
      style="background: var(--highlighted-card-bg-color)"
      v-if="item.template_id"
      class="mb-8 pt-3"
    >
      <div style="
        position: absolute;
        background: var(--highlighted-card-bg-color);
        width: 28px;
        height: 28px;
        transform: rotate(45deg);
        left: calc(50% - 14px);
        top: -14px;
        border-radius: 0;
      "></div>

      <v-card-text>
        <TaskParamsForm
          :template="templates.find(t => t.id === item.template_id)"
          v-model="item.task_params"
        />

      </v-card-text>
    </v-card>

    <div v-if="type === 'run_at'">
      <v-text-field

        v-model="runAtInput"
        type="datetime-local"
        label="Run at"
        :rules="runAtRules"
        :disabled="formSaving"
        :suffix="timezone + ' time'"
        outlined
        dense
      ></v-text-field>

      <div class="d-flex justify-end">

        <v-checkbox
          v-model="item.delete_after_run"
          hide-details
          class="mt-0 pt-0"
        >
          <template v-slot:label>
            {{ $t('Delete after run') }}
          </template>
        </v-checkbox>
      </div>
    </div>

    <div v-else>
      <v-switch
        v-model="rawCron"
        label="Show cron format"
        :disabled="disableRawCron"
      />

      <template v-if="rawCron">
        <v-text-field
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

        <ScheduleOffsetField
          v-model="item.offset_days"
          :label="$t('scheduleOffset')"
          :hint="offsetHint"
          :disabled="formSaving"
          @input="refreshCheckboxes()"
        />
      </template>

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
          @change="selectTiming()"
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

        <div v-if="timing === 'monthly_weekday'">
          <div class="mt-4">{{ $t('scheduleWeekOfMonth') }}</div>
          <div class="d-flex flex-wrap">
            <v-checkbox
              class="mr-2 mt-0 ScheduleCheckbox"
              v-for="o in ORDINALS"
              :key="o.id"
              :value="o.id"
              :label="o.title"
              :input-value="ordinals"
              color="white"
              :class="{'ScheduleCheckbox--active': ordinals.includes(o.id)}"
              @change="(values) => selectOrdinals(values, o.id)"
            ></v-checkbox>
          </div>
        </div>

        <div v-if="['weekly', 'monthly_weekday'].includes(timing)">
          <div class="mt-4">Weekdays</div>
          <div class="d-flex flex-wrap">
            <v-checkbox
              class="mr-2 mt-0 ScheduleCheckbox"
              v-for="d in WEEKDAYS" :key="d.id"
              :value="d.id"
              :label="d.title"
              :input-value="weekdays"
              color="white"
              :class="{'ScheduleCheckbox--active': weekdays.includes(d.id)}"
              @change="(values) => selectWeekdays(values, d.id)"
            ></v-checkbox>
          </div>
        </div>

        <div v-if="timing === 'monthly_weekday'">
          <div class="mt-4 mb-2">{{ $t('scheduleOffset') }}</div>
          <ScheduleOffsetField
            v-model="item.offset_days"
            :hint="offsetHint"
            :disabled="formSaving"
            @input="validateCronFormat()"
          />
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

        <div v-if="['yearly', 'monthly', 'monthly_weekday', 'weekly', 'daily'].includes(timing)">
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

        <v-alert
          v-if="cronFormatError"
          class="mt-4 mb-0"
          type="error"
          text
          dense
        >{{ cronFormatError }}</v-alert>
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

    <div
      v-if="upcomingRuns.length > 0"
      class="d-flex flex-wrap align-center mb-4"
    >
      <span class="text-caption mr-2">{{ $t('scheduleThen') }}</span>
      <v-chip
        v-for="run in upcomingRuns"
        :key="run.key"
        class="mr-1 my-1"
        small
        outlined
      >{{ run.label }}</v-chip>
    </div>

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
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezonePlugin from 'dayjs/plugin/timezone';
import customParseFormat from 'dayjs/plugin/customParseFormat';

import { CronExpressionParser } from 'cron-parser';
import {
  isWeekly,
  isYearly,
  isMonthly,
  isDaily,
  isHourly,
  pruneSelectionsForTiming,
  buildCronFormat,
  usesQuartzDays,
  parseWeekdayOfMonth,
  buildWeekdayOfMonthCronFormat,
  LAST,
} from '@/lib/cronPresets';
import { getErrorMessage } from '@/lib/error';
import TaskParamsForm from '@/components/TaskParamsForm.vue';
import ScheduleOffsetField from '@/components/ScheduleOffsetField.vue';

dayjs.extend(utc);
dayjs.extend(timezonePlugin);
dayjs.extend(customParseFormat);

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
  id: 'monthly_weekday',
  title: 'Monthly by weekday',
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

const ORDINALS = [
  { id: 1, title: '1st' },
  { id: 2, title: '2nd' },
  { id: 3, title: '3rd' },
  { id: 4, title: '4th' },
  { id: LAST, title: 'Last' },
];

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

const RUN_AT_FORMAT = 'YYYY-MM-DDTHH:mm';

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

function formatRunInTZ(date, tz) {
  const weekday = new Intl.DateTimeFormat('en-GB', { timeZone: tz, weekday: 'short' }).format(date);
  return `${weekday} ${formatDateInTZ(date, tz)} ${formatTimeInTZ(date, tz)}`;
}

export default {
  components: { TaskParamsForm, ScheduleOffsetField },
  mixins: [ItemFormBase],

  data() {
    return {
      templates: null,
      timing: 'hourly',
      TIMINGS,
      MONTHS,
      WEEKDAYS,
      ORDINALS,
      MINUTES,
      minutes: [],
      hours: [],
      days: [],
      months: [],
      weekdays: [],
      ordinals: [],
      nextRuns: [],
      validationRequest: 0,
      rawCron: false,
      disableRawCron: false,
      showInfo: true,
      cronFormatError: null,
      runAtInput: '',
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
    this.showInfo = localStorage.getItem('schedule__hide_info') !== '1';
    this.rawCron = localStorage.getItem('schedule__raw_cron') !== '1';

    this.templates = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/templates`,
      responseType: 'json',
    })).data;
  },

  props: {
    timezone: String,
    type: String,
  },

  computed: {
    localTimezone() {
      return 'Local';
    },

    runAtRules() {
      if (this.type === 'run_at') {
        return [];
      }

      return [
        (v) => !!v || 'Run time is required',
      ];
    },

    upcomingRuns() {
      return this.nextRuns.slice(1).map((run) => ({
        key: run.toISOString(),
        label: formatRunInTZ(run, this.timezone),
      }));
    },

    offsetHint() {
      const days = this.item.offset_days;

      if (!days) {
        return this.$t('scheduleOffsetNone');
      }

      const count = Math.abs(days);
      return this.$tc(days > 0 ? 'scheduleOffsetAfter' : 'scheduleOffsetBefore', count, { count });
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
    // The form is reused between schedules, so drop the previous one's preview
    // and any check still in flight for it.
    beforeLoadData() {
      this.validationRequest += 1;
      this.nextRuns = [];
      this.cronFormatError = null;
    },

    getNewItem() {
      return {
        name: '',
        template_id: null,
        cron_format: '* * * * *',
        offset_days: 0,
        active: true,
        run_once: false,
        delete_after_run: false,
        task_params: {},
        run_at: null,
      };
    },

    setDefaultRunAt() {
      const nextHour = dayjs().tz(this.timezone).add(1, 'hour').minute(0)
        .second(0)
        .millisecond(0);

      this.runAtInput = nextHour.format(RUN_AT_FORMAT);
    },

    setRunAtInputFromItem() {
      if (!this.item.run_at) {
        this.runAtInput = '';
        return;
      }

      const parsed = dayjs(this.item.run_at).tz(this.timezone);
      this.runAtInput = parsed.isValid() ? parsed.format(RUN_AT_FORMAT) : '';
    },

    nextRunTime() {
      if (this.type === 'run_at') {
        const runAt = this.item.run_at ? dayjs(this.item.run_at) : null;
        const parsed = this.runAtInput
          ? dayjs.tz(this.runAtInput, RUN_AT_FORMAT, this.timezone)
          : runAt;

        if (!parsed || !parsed.isValid()) {
          return null;
        }

        return parsed.toDate();
      }

      return this.nextRuns[0] || null;
    },

    // Checks the schedule on the server, which also returns the next runs, so
    // the preview always matches what the scheduler will do. Returns the error
    // message, null when valid, or undefined when a newer check superseded it.
    async validateCronFormat() {
      this.validationRequest += 1;
      const request = this.validationRequest;
      this.nextRuns = [];

      let nextRuns = [];
      let error = null;

      try {
        const res = await axios({
          method: 'post',
          url: `/api/project/${this.projectId}/schedules/validate`,
          responseType: 'json',
          data: {
            project_id: this.projectId,
            cron_format: this.item.cron_format,
            offset_days: this.item.offset_days,
          },
        });
        nextRuns = (res.data.next_runs || []).map((run) => new Date(run));
      } catch (err) {
        error = getErrorMessage(err);
      }

      if (request !== this.validationRequest) {
        return undefined;
      }

      this.nextRuns = nextRuns;
      this.cronFormatError = error;
      return error;
    },

    async refreshCheckboxes() {
      if (this.type === 'run_at') {
        this.cronFormatError = null;
        this.disableRawCron = false;
        return;
      }

      // if (!/test/.test(this.item.cron_format)) {
      //   this.rawCron = true;
      //   this.disableRawCron = true;
      // } else {
      //   this.disableRawCron = false;
      // }

      const cronError = await this.validateCronFormat();

      if (cronError === undefined) {
        return; // the value changed while validating, ignore stale result
      }

      const weekdayOfMonth = parseWeekdayOfMonth(this.item.cron_format);

      // Only the "monthly by weekday" timing can show Quartz days and offsets.
      const fitsTimings = weekdayOfMonth != null
        || (!this.item.offset_days && !usesQuartzDays(this.item.cron_format));

      if (cronError != null || !fitsTimings) {
        this.rawCron = true;
        this.disableRawCron = true;
        return;
      }

      this.disableRawCron = false;

      const [minuteField, hourField] = this.item.cron_format.trim().split(/\s+/);

      // cron-parser rejects lists of # items, so for those it reads only the time.
      const expression = weekdayOfMonth
        ? `${minuteField} ${hourField} * * *`
        : this.item.cron_format;

      let cron;
      try {
        cron = CronExpressionParser.parse(expression, {
          tz: this.timezone,
        });
      } catch {
        // Valid on the backend but not parseable by cron-parser
        // (e.g. @hourly) — show it in raw mode without an error.
        this.rawCron = true;
        this.disableRawCron = true;
        return;
      }

      const fields = cron.fields; // JSON.parse(JSON.stringify(cron.fields));

      if (weekdayOfMonth) {
        this.timing = 'monthly_weekday';
        this.months = [];
        this.days = [];
        this.weekdays = weekdayOfMonth.weekdays;
        this.ordinals = weekdayOfMonth.ordinals;
        this.hours = hourField === '*' ? [] : fields.hour.values;
        this.minutes = minuteField === '*' ? [] : fields.minute.values;
        return;
      }

      // The form is reused between schedules, so "* * * * *", which matches
      // none of the checks below, must not keep the previous timing.
      this.timing = 'hourly';
      this.ordinals = [];
      this.months = [];
      this.weekdays = [];
      this.hours = [];
      this.minutes = [];

      if (this.isHourly(this.item.cron_format)) {
        this.minutes = fields.minute.values;
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
      // if (!this.item.type) {
      //   this.item.type = this.item.run_at ? 'run_at' : '';
      // }

      if (this.item.run_at) {
        this.setRunAtInputFromItem();
      } else if (this.type === 'run_at') {
        this.setDefaultRunAt();
      } else if (this.isNew) {
        this.item.cron_format = '* * * * *';
      }

      this.refreshCheckboxes();
    },

    async beforeSave() {
      this.item.type = this.type;

      if (this.type === 'run_at') {
        const parsed = this.runAtInput
          ? dayjs.tz(this.runAtInput, RUN_AT_FORMAT, this.timezone)
          : null;

        if (!parsed || !parsed.isValid()) {
          this.formError = 'Please provide a valid run time for the run_at schedule.';
          throw new Error(this.formError);
        }

        this.item.run_at = parsed.toISOString();
        this.item.cron_format = this.item.cron_format || '';
      } else {
        this.item.run_at = null;
      }
    },

    isWeekly(s) {
      return isWeekly(s);
    },

    isYearly(s) {
      return isYearly(s);
    },

    isMonthly(s) {
      return isMonthly(s);
    },

    isDaily(s) {
      return isDaily(s);
    },

    isHourly(s) {
      return isHourly(s);
    },

    selectTiming() {
      if (this.timing === 'monthly_weekday') {
        // The day of week needs at least one weekday and one ordinal.
        this.weekdays = this.weekdays.length > 0 ? this.weekdays : [1];
        this.ordinals = this.ordinals.length > 0 ? this.ordinals : [1];
      }

      this.refreshCron();
    },

    // Unticking the last weekday or ordinal of a monthly-by-weekday schedule
    // keeps it ticked; a weekly schedule without weekdays runs every day.
    selectWeekdays(values, weekday) {
      this.weekdays = values.length === 0 && this.timing === 'monthly_weekday'
        ? [weekday]
        : values;
      this.refreshCron();
    },

    selectOrdinals(values, ordinal) {
      this.ordinals = values.length === 0 ? [ordinal] : values;
      this.refreshCron();
    },

    refreshCron() {
      const selections = pruneSelectionsForTiming(this.timing, {
        months: this.months,
        weekdays: this.weekdays,
        days: this.days,
        hours: this.hours,
        minutes: this.minutes,
      });

      this.months = selections.months;
      this.weekdays = selections.weekdays;
      this.days = selections.days;
      this.hours = selections.hours;
      this.minutes = selections.minutes;

      if (this.timing === 'monthly_weekday') {
        this.item.cron_format = buildWeekdayOfMonthCronFormat({
          ...selections,
          ordinals: this.ordinals,
        });
      } else {
        this.item.cron_format = buildCronFormat(selections);
        this.item.offset_days = 0;
      }

      this.validateCronFormat();
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
