async function getDomains() {
  const response = await fetch("/api/domains");
  const data = await response.json();
  return data as string[];
}

async function getRecords(domain: string) {
  const response = await fetch(`/api/records/` + domain);
  const data = await response.json();
  return data as record[];
}

async function addRecord(
  domain: string,
  type: RecordType,
  dest: string,
  port: number,
  waf_enabled: boolean,
  proxy: boolean,
) {
  const response = await fetch(`/api/records/add`, {
    method: "POST",
    body: JSON.stringify({
      domain: domain,
      type: type,
      dest: dest,
      port: port,
      waf_enabled: waf_enabled,
      proxy: proxy,
    }),
  });
  return (response.status == 200) as boolean;
}

async function removeRecord(id: number, domain: string) {
  const response = await fetch(`/api/records/remove`, {
    method: "POST",
    body: JSON.stringify({
      id: id,
      domain: domain,
    }),
  });
  return (response.status == 200) as boolean;
}

class record {
  cad: CadEntry;
  dns: DNSEntry;
}

class CadEntry {
  domain: string;
  dest: string;
  port: number;
  waf: boolean;
}

class DNSEntry {
  id: number;
  domain: string;
  type: number;
  value: string;
}

// Enum for record types
enum RecordType {
  A = "A",
  AAAA = "AAAA",
  CNAME = "CNAME",
  MX = "MX",
  TXT = "TXT",
  NS = "NS",
  SRV = "SRV",
}
