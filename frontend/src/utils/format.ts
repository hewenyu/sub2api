export function formatDate(dateString: string | null | undefined): string {
  if (!dateString) return '-';

  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return '-';

    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    }).format(date);
  } catch {
    return '-';
  }
}

export function formatDateTime(dateString: string | null | undefined): string {
  return formatDate(dateString);
}

export function formatNumber(num: number | null | undefined): string {
  if (num === null || num === undefined) return '0';

  return new Intl.NumberFormat('en-US').format(num);
}

export function formatBytes(bytes: number | null | undefined): string {
  if (!bytes || bytes === 0) return '0 B';

  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

export function formatTokens(tokens: number | null | undefined): string {
  if (tokens === null || tokens === undefined) return '0';

  if (tokens >= 1000000) {
    return `${(tokens / 1000000).toFixed(2)}M`;
  } else if (tokens >= 1000) {
    return `${(tokens / 1000).toFixed(2)}K`;
  }

  return tokens.toString();
}

export function truncateString(str: string, maxLength: number): string {
  if (!str || str.length <= maxLength) return str;

  return str.substring(0, maxLength) + '...';
}

export function maskApiKey(apiKey: string): string {
  if (!apiKey || apiKey.length < 8) return apiKey;

  const start = apiKey.substring(0, 4);
  const end = apiKey.substring(apiKey.length - 4);

  return `${start}...${end}`;
}

export function formatPercentage(value: number, total: number): string {
  if (!total || total === 0) return '0%';

  const percentage = (value / total) * 100;

  return `${percentage.toFixed(2)}%`;
}

export function formatCost(cost: number | null | undefined): string {
  if (cost === null || cost === undefined || cost === 0) return '$0.00';

  return `$${cost.toFixed(4)}`;
}
