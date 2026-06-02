// NOTE: Narrow regression guard for CLA-33 until the frontend has component tests.
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import assert from 'node:assert/strict';

import {
  DEFAULT_LEASE_DURATION_SECONDS,
  MAX_LEASE_DURATION_SECONDS,
  validatePortMapping,
} from '../src/lib/validate.ts';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const backendModels = readFileSync(resolve(root, '../backend/internal/domain/models.go'), 'utf8');

assert(
  backendModels.includes('const MaxLeaseDurationSeconds = 7 * 24 * 60 * 60'),
  'backend MaxLeaseDurationSeconds expression changed; update frontend contract intentionally',
);
assert.equal(MAX_LEASE_DURATION_SECONDS, 604800, 'frontend max lease duration must remain 7 days');
assert.equal(DEFAULT_LEASE_DURATION_SECONDS, 7200, 'new port form default lease must remain 2 hours');
assert(
  DEFAULT_LEASE_DURATION_SECONDS > 0 && DEFAULT_LEASE_DURATION_SECONDS <= MAX_LEASE_DURATION_SECONDS,
  'default lease duration must be a finite lease within the allowed range',
);

const validMapping = {
  protocol: 'TCP',
  external_port: 65535,
  internal_ip: '',
  internal_port: 65535,
  description: 'contract check',
  lease_duration_seconds: 604800,
};

assert.deepEqual(
  validatePortMapping(validMapping, { allowAutoInternalIP: true }),
  [],
  '65535 ports and 604800 second leases must be accepted for auto-resolved internal IPs',
);
assert(
  validatePortMapping({ ...validMapping, external_port: Number.NaN }, { allowAutoInternalIP: true })
    .some((message) => message.includes('external_port')),
  'NaN external ports must be rejected',
);
assert(
  validatePortMapping({ ...validMapping, external_port: Number.parseInt('', 10) }, { allowAutoInternalIP: true })
    .some((message) => message.includes('external_port')),
  'empty port input must be rejected after parsing',
);
assert(
  validatePortMapping({ ...validMapping, external_port: 65536 }, { allowAutoInternalIP: true })
    .some((message) => message.includes('external_port')),
  '65536 external ports must be rejected',
);
assert(
  validatePortMapping({ ...validMapping, internal_port: 65536 }, { allowAutoInternalIP: true })
    .some((message) => message.includes('internal_port')),
  '65536 internal ports must be rejected',
);
assert(
  validatePortMapping({ ...validMapping, lease_duration_seconds: 604801 }, { allowAutoInternalIP: true })
    .some((message) => message.includes('lease_duration_seconds')),
  '604801 second leases must be rejected',
);
assert(
  validatePortMapping(validMapping).some((message) => message.includes('internal_ip')),
  'strict validation must still require internal_ip',
);
