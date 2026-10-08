import './setup';
import { expect } from 'chai';
import { shallowMount } from '@vue/test-utils';
import ScheduleForm from '@/components/ScheduleForm.vue';
import ScheduleOffsetField from '@/components/ScheduleOffsetField.vue';
import mockAxios, { httpError } from './helpers/axiosMock';

const FormStub = {
  render(h) {
    return h('form', this.$slots.default);
  },
  methods: {
    validate: () => true,
    resetValidation: () => {},
  },
};

const SCHEDULES = {
  2: {
    id: 2, type: '', cron_format: '0 3 * * 2#2', offset_days: 1, task_params: {},
  },
  3: {
    id: 3, type: '', cron_format: '0 3 * * *', offset_days: 0, task_params: {},
  },
  5: {
    id: 5, type: 'run_at', run_at: '2030-01-01T10:00:00Z', cron_format: '', offset_days: 0, task_params: {},
  },
};

const RUNS = ['2026-10-14T01:00:00Z', '2026-11-11T02:00:00Z', '2026-12-09T02:00:00Z'];

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

function mountForm(propsData) {
  return shallowMount(ScheduleForm, {
    propsData: {
      projectId: 1, timezone: 'Europe/Stockholm', type: '', ...propsData,
    },
    mocks: { $t: (key) => key, $tc: (key) => key },
    stubs: { 'v-form': FormStub },
  });
}

// Other specs replace localStorage with partial stubs, so this one brings its own.
function useLocalStorage(store) {
  Object.defineProperty(global, 'localStorage', {
    configurable: true,
    writable: true,
    value: store,
  });
}

describe('ScheduleForm.vue', () => {
  const originalLocalStorage = global.localStorage;
  let http;
  let validate;

  after(() => {
    useLocalStorage(originalLocalStorage);
  });

  beforeEach(() => {
    const items = new Map();
    useLocalStorage({
      getItem: (key) => (items.has(key) ? items.get(key) : null),
      setItem: (key, value) => items.set(key, String(value)),
      removeItem: (key) => items.delete(key),
    });

    validate = () => ({ next_runs: RUNS });
    http = mockAxios();
    http.respond((config) => {
      if (config.url.endsWith('/templates')) {
        return [{ id: 1, name: 'Deploy' }];
      }
      if (config.url.endsWith('/schedules/validate')) {
        return validate(config.data);
      }
      // A copy, because the form edits the item it loads.
      return { ...SCHEDULES[config.url.split('/').pop()] };
    });
  });

  afterEach(() => {
    http.restore();
  });

  it('previews the runs the server computes, offset included', async () => {
    const wrapper = mountForm({ itemId: 2 });
    await flush();

    const requests = http.requests.filter((r) => r.url.endsWith('/schedules/validate'));
    expect(requests[requests.length - 1].data).to.include({ cron_format: '0 3 * * 2#2', offset_days: 1 });
    expect(wrapper.vm.nextRuns.map((run) => run.getTime()))
      .to.deep.equal(RUNS.map((run) => Date.parse(run)));
    expect(wrapper.vm.upcomingRuns.map((run) => run.label))
      .to.deep.equal(['Wed 2026-11-11 03:00', 'Wed 2026-12-09 03:00']);
  });

  it('ignores a check that a newer one superseded', async () => {
    const wrapper = mountForm({ itemId: 3 });
    await flush();

    let release;
    validate = () => new Promise((resolve) => { release = resolve; });
    wrapper.vm.item.cron_format = '0 4 * * *';
    const stale = wrapper.vm.refreshCheckboxes();
    const resolveStale = release;

    validate = () => ({ next_runs: ['2026-10-09T02:30:00Z'] });
    wrapper.vm.item.cron_format = '30 4 * * *';
    await wrapper.vm.refreshCheckboxes();

    resolveStale({ next_runs: RUNS });
    await stale;

    expect(wrapper.vm.nextRuns.map((run) => run.toISOString()))
      .to.deep.equal(['2026-10-09T02:30:00.000Z']);
  });

  it('starts each schedule without the previous one\'s preview or error', async () => {
    const wrapper = mountForm({ itemId: 3 });
    await flush();

    validate = () => { throw httpError(400, { error: 'Cron: bad' }); };
    wrapper.vm.item.cron_format = '0 3 * * x';
    await wrapper.vm.refreshCheckboxes();
    expect(wrapper.vm.cronFormatError).to.equal('Cron: bad');

    await wrapper.setProps({ itemId: 5, type: 'run_at' });
    await wrapper.vm.reset();
    await flush();

    expect(wrapper.vm.cronFormatError).to.equal(null);
    expect(wrapper.vm.upcomingRuns).to.deep.equal([]);
  });

  it('keeps raw mode while an offset is set on a format the timings cannot show', async () => {
    const wrapper = mountForm({ itemId: 3 });
    await flush();
    expect(wrapper.vm.disableRawCron).to.equal(false);

    const offsetField = () => wrapper.findComponent(ScheduleOffsetField);

    offsetField().vm.$emit('input', 1);
    await flush();
    expect(wrapper.vm.disableRawCron).to.equal(true);

    offsetField().vm.$emit('input', 0);
    await flush();
    expect(wrapper.vm.disableRawCron).to.equal(false);
  });

  it('keys the runs by time, so a repeated hour does not clash', async () => {
    const wrapper = mountForm({ itemId: 3 });
    await flush();

    wrapper.vm.nextRuns = ['2026-10-24T23:30:00Z', '2026-10-25T00:30:00Z', '2026-10-25T01:30:00Z']
      .map((run) => new Date(run));

    const keys = wrapper.vm.upcomingRuns.map((run) => run.key);
    expect(new Set(keys).size).to.equal(keys.length);
  });
});
