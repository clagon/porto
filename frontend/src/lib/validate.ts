import type { PortMapping, Settings } from './types';

export const MAX_PORT_NUMBER = 65535;
export const MAX_LEASE_DURATION_SECONDS = 7 * 24 * 60 * 60;

type PortMappingValidationOptions = {
  allowAutoInternalIP?: boolean;
};

export function validateSettings(settings: Settings): string[] {
  const errors: string[] = [];
  const addr = settings.listen_addr.trim();
  if (!addr) {
    errors.push('listen_addr is required');
  }
  if (!/^(127\.0\.0\.1|localhost|\[::1\]):\d+$/.test(addr)) {
    errors.push('listen_addr must stay on localhost');
  }
  return errors;
}

export function validatePortMapping(mapping: PortMapping, options: PortMappingValidationOptions = {}): string[] {
  const errors: string[] = [];
  const protocol = mapping.protocol.trim().toUpperCase();
  if (protocol !== 'TCP' && protocol !== 'UDP') {
    errors.push('protocol must be TCP or UDP');
  }
  if (!Number.isInteger(mapping.external_port) || mapping.external_port < 1 || mapping.external_port > MAX_PORT_NUMBER) {
    errors.push(`external_port must be 1-${MAX_PORT_NUMBER}`);
  }
  if (!options.allowAutoInternalIP && !mapping.internal_ip.trim()) {
    errors.push('internal_ip is required');
  }
  if (!Number.isInteger(mapping.internal_port) || mapping.internal_port < 1 || mapping.internal_port > MAX_PORT_NUMBER) {
    errors.push(`internal_port must be 1-${MAX_PORT_NUMBER}`);
  }
  if (!Number.isInteger(mapping.lease_duration_seconds) || mapping.lease_duration_seconds < 0) {
    errors.push('lease_duration_seconds must be 0 or greater');
  }
  if (Number.isInteger(mapping.lease_duration_seconds) && mapping.lease_duration_seconds > MAX_LEASE_DURATION_SECONDS) {
    errors.push(`lease_duration_seconds must be ${MAX_LEASE_DURATION_SECONDS} or less`);
  }
  return errors;
}

export function validateCloseMapping(mapping: PortMapping): string[] {
  const errors: string[] = [];
  const protocol = mapping.protocol.trim().toUpperCase();
  if (protocol !== 'TCP' && protocol !== 'UDP') {
    errors.push('protocol must be TCP or UDP');
  }
  if (!Number.isInteger(mapping.external_port) || mapping.external_port < 1 || mapping.external_port > MAX_PORT_NUMBER) {
    errors.push(`external_port must be 1-${MAX_PORT_NUMBER}`);
  }
  return errors;
}
