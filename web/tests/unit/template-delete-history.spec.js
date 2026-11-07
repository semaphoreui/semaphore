import './setup';
import { expect } from 'chai';
import EventBus from '@/event-bus';
import TemplateView from '@/views/project/TemplateView.vue';
import mockAxios, { httpError } from './helpers/axiosMock';

describe('TemplateView.vue task history deletion', () => {
  let http;
  let snackbars;
  const onSnackbar = (e) => snackbars.push(e);

  beforeEach(() => {
    http = mockAxios();
    snackbars = [];
    EventBus.$on('i-snackbar', onSnackbar);
  });

  afterEach(() => {
    http.restore();
    EventBus.$off('i-snackbar', onSnackbar);
  });

  const ctx = () => ({
    projectId: 3,
    itemId: 7,
    deleteHistoryDialog: true,
    deleteHistoryProgress: false,
    needLoadData: false,
    $t: (key) => key,
  });

  it('is offered only to an admin, as the server allows', () => {
    expect(TemplateView.computed.canDeleteHistory.call({ isAdmin: true })).to.equal(true);
    expect(TemplateView.computed.canDeleteHistory.call({ isAdmin: false })).to.equal(false);
  });

  it('deletes the history of the template and reloads the task list', async () => {
    const c = ctx();

    await TemplateView.methods.deleteTaskHistory.call(c);

    expect(http.requests).to.have.length(1);
    expect(http.requests[0].method).to.equal('delete');
    expect(http.requests[0].url).to.equal('/api/project/3/templates/7/tasks');
    expect(c.needLoadData).to.equal(true);
    expect(c.deleteHistoryDialog).to.equal(false);
    expect(c.deleteHistoryProgress).to.equal(false);
    expect(snackbars).to.deep.equal([{ color: 'success', text: 'taskHistoryDeleted' }]);
  });

  [
    ['a refused request', 401],
    ['a failure after some batches were deleted', 500],
  ].forEach(([name, status]) => {
    it(`reports ${name} and still reloads the task list`, async () => {
      http.respond(() => { throw httpError(status, { error: 'Failed' }); });
      const c = ctx();

      await TemplateView.methods.deleteTaskHistory.call(c);

      expect(c.needLoadData).to.equal(true);
      expect(c.deleteHistoryDialog).to.equal(false);
      expect(c.deleteHistoryProgress).to.equal(false);
      expect(snackbars).to.have.length(1);
      expect(snackbars[0].color).to.equal('error');
    });
  });
});
