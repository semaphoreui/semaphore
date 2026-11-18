import './setup';
import {
  describe, it, beforeEach, afterEach, vi,
} from 'vitest';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import i18n from '@/plugins/i18';
import ViewTemplateStats from '@/components/ViewTemplateStats.vue';
import Templates from '@/views/project/Templates.vue';
import EditViewsForm from '@/components/EditViewsForm.vue';
import mockAxios from './helpers/axiosMock';

describe('ViewTemplateStats', () => {
  function segments(stats) {
    const wrapper = mount(ViewTemplateStats, {
      props: { stats },
      global: { plugins: [i18n] },
    });
    const res = wrapper.findAll('.view-template-stats__segment').map((chip) => ({
      testid: chip.attributes('data-testid'),
      text: chip.text(),
      title: chip.attributes('title'),
    }));
    wrapper.unmount();
    return res;
  }

  it('draws a segment per count, in the order success, failed, stopped, confirming, running', () => {
    expect(segments({
      running: 5, confirming: 4, stopped: 3, failed: 2, success: 1,
    })).to.deep.equal([
      { testid: 'viewTemplateStats-success', text: '1', title: 'Success: 1' },
      { testid: 'viewTemplateStats-failed', text: '2', title: 'Failed: 2' },
      { testid: 'viewTemplateStats-stopped', text: '3', title: 'Stopped / Not launched: 3' },
      { testid: 'viewTemplateStats-confirming', text: '4', title: 'Waiting confirmation: 4' },
      { testid: 'viewTemplateStats-running', text: '5', title: 'Running: 5' },
    ]);
  });

  it('leaves out the counts of zero', () => {
    expect(segments({
      success: 5, failed: 0, stopped: 0, confirming: 1,
    }).map((s) => s.testid)).to.deep.equal([
      'viewTemplateStats-success',
      'viewTemplateStats-confirming',
    ]);
  });

  it('draws nothing without counts', () => {
    [undefined, null, {}, {
      success: 0, failed: 0, stopped: 0, confirming: 0, running: 0,
    }].forEach((stats) => {
      const wrapper = mount(ViewTemplateStats, { props: { stats }, global: { plugins: [i18n] } });
      expect(wrapper.find('[data-testid="viewTemplateStats"]').exists()).to.equal(false);
      wrapper.unmount();
    });
  });

  it('colours the segments as the task statuses', () => {
    const wrapper = mount(ViewTemplateStats, {
      props: {
        stats: {
          success: 1, failed: 1, stopped: 1, confirming: 1, running: 1,
        },
      },
      global: { plugins: [i18n] },
    });
    // The chip takes its variant from the defaults of the app (flat: `bg-`; the bare test
    // instance is tonal: `text-`), the colour is the same either way.
    const colour = (key) => wrapper.find(`[data-testid="viewTemplateStats-${key}"]`).classes()
      .map((c) => (/^(?:bg|text)-(success|error|warning|primary)$/.exec(c) || [])[1])
      .find(Boolean);
    expect(colour('success')).to.equal('success');
    expect(colour('failed')).to.equal('error');
    expect(colour('confirming')).to.equal('warning');
    expect(colour('running')).to.equal('primary');
    expect(colour('stopped')).to.equal(undefined);
    wrapper.unmount();
  });
});

// The methods of the Templates page that keep the counts of the tabs fresh, run on a bare
// context: mounting the page needs the whole project.
describe('Templates view stats refresh', () => {
  let http;

  function context(views) {
    return {
      projectId: 7,
      views,
      viewStatsTimer: null,
      getViewsUrl: Templates.methods.getViewsUrl,
      refreshViewStats: Templates.methods.refreshViewStats,
      scheduleViewStatsRefresh: Templates.methods.scheduleViewStatsRefresh,
    };
  }

  beforeEach(() => {
    http = mockAxios();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    http.restore();
  });

  it('asks the views with the template counts', () => {
    expect(context([]).getViewsUrl()).to.equal('/api/project/7/views?include=template_stats');
  });

  it('updates the counts of the shown views and keeps the views themselves', async () => {
    const shown = [
      {
        id: 1, title: 'All', show_template_stats: true, template_stats: { success: 1 },
      },
      {
        id: 2, title: 'Build', show_template_stats: true, template_stats: { success: 1 },
      },
    ];
    const ctx = context(shown);
    http.respond(() => [
      { id: 1, title: 'renamed', template_stats: { success: 3, failed: 1 } },
      // a hidden view the page left out
      { id: 9, title: 'Hidden', template_stats: { success: 9 } },
    ]);

    await ctx.refreshViewStats();

    expect(ctx.views).to.equal(shown);
    expect(shown).to.deep.equal([
      {
        id: 1, title: 'All', show_template_stats: true, template_stats: { success: 3, failed: 1 },
      },
      {
        id: 2, title: 'Build', show_template_stats: true, template_stats: { success: 1 },
      },
    ]);
  });

  it('joins a burst of task events into one request', async () => {
    const ctx = context([{ id: 1, show_template_stats: true, template_stats: {} }]);
    http.respond(() => [{ id: 1, template_stats: { failed: 2 } }]);

    ctx.scheduleViewStatsRefresh();
    ctx.scheduleViewStatsRefresh();
    ctx.scheduleViewStatsRefresh();
    expect(http.requests).to.have.length(0);

    await vi.advanceTimersByTimeAsync(1000);

    expect(http.requests).to.have.length(1);
    expect(ctx.views[0].template_stats).to.deep.equal({ failed: 2 });

    // the next event after the request starts a new one
    ctx.scheduleViewStatsRefresh();
    await vi.advanceTimersByTimeAsync(1000);
    expect(http.requests).to.have.length(2);
  });

  it('survives a failed request and retries on the next event', async () => {
    const ctx = context([{ id: 1, show_template_stats: true, template_stats: { success: 1 } }]);
    http.respond(() => { throw new Error('network'); });

    ctx.scheduleViewStatsRefresh();
    await vi.advanceTimersByTimeAsync(1000);

    expect(ctx.views[0].template_stats).to.deep.equal({ success: 1 });
    expect(ctx.viewStatsTimer).to.equal(null);

    http.respond(() => [{ id: 1, template_stats: { success: 2 } }]);
    ctx.scheduleViewStatsRefresh();
    await vi.advanceTimersByTimeAsync(1000);
    expect(ctx.views[0].template_stats).to.deep.equal({ success: 2 });
  });

  it('asks nothing when no view shows the counts', async () => {
    const ctx = context([{ id: 1, title: 'All' }, { id: 2, title: 'Build' }]);

    ctx.scheduleViewStatsRefresh();
    await vi.advanceTimersByTimeAsync(5000);

    expect(ctx.viewStatsTimer).to.equal(null);
    expect(http.requests).to.have.length(0);
  });
});

// The setting of a view that turns the counts on: off unless ticked, saved with the view.
describe('EditViewsForm show task counts', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
  });

  afterEach(() => {
    http.restore();
  });

  async function save(view) {
    const ctx = { projectId: 7, views: [view] };
    http.respond(() => ({ id: 5 }));
    await EditViewsForm.methods.saveView.call(ctx, view.id);
    return http.requests[http.requests.length - 1];
  }

  it('sends the setting of an existing view', async () => {
    const request = await save({
      id: 3, title: 'Build', type: '', show_template_stats: true,
    });
    expect(request.method).to.equal('put');
    expect(request.url).to.equal('/api/project/7/views/3');
    expect(request.data.show_template_stats).to.equal(true);
  });

  it('leaves the setting off for a new view', async () => {
    const request = await save({ id: -1, title: 'New', type: '' });
    expect(request.method).to.equal('post');
    expect(request.data).to.not.have.property('show_template_stats');
  });

  it('has the checkbox in the settings of a view', () => {
    const source = EditViewsForm.render ? EditViewsForm.render.toString() : '';
    expect(source).to.contain('viewsForm-showTemplateStats');
    expect(source).to.contain('Show task counts');
  });
});
