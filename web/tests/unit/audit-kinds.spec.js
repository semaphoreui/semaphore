import { expect } from 'chai';
import en from '@/lang/en';
import AUDIT_KINDS from '@/lib/auditKinds';

describe('lib/auditKinds', () => {
  it('has an English title for every kind', () => {
    const missing = AUDIT_KINDS.filter((kind) => !en[`audit_kind_${kind.replace(/[./]/g, '_')}`]);
    expect(missing).to.deep.equal([]);
  });
});

describe('AuditFilters.vue kind filter', () => {
  it('offers every kind', () => {
    // eslint-disable-next-line global-require
    const AuditFilters = require('@/components/AuditFilters.vue').default;
    const i18n = { t: (k) => k, te: () => true };
    const items = AuditFilters.computed.kindItems.call({ $t: (k) => k, $i18n: i18n });
    const offered = items.filter((i) => i.value).map((i) => i.value);
    const missing = AUDIT_KINDS.filter((kind) => !offered.includes(kind));
    expect(missing).to.deep.equal([]);
  });
});
