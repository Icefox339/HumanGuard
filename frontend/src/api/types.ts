export type Site = {
  id: string;
  domain: string;
  status: 'active' | 'paused';
};

export type Session = {
  id: string;
  ip: string;
  riskScore: number;
};

export type User = {
  id: string;
  email: string;
  name?: string;
  role: 'user' | 'admin';
};
