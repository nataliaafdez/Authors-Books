import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

const client = new grpc.Client();
client.load(
  ['/Users/natalia/AuthorsBooks/proto/authorsv1'],
  'authors.proto'
);

export const options = {
  stages: [
    { duration: '10s', target: 10 },
    { duration: '20s', target: 30 },
    { duration: '20s', target: 50 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    checks: ['rate>0.95'], 
  },
};

export default () => {
  const addr = __ENV.AUTHORS_ADDR;
  if (!addr) {
    throw new Error('AUTHORS_ADDR no está definido');
  }

  client.connect(addr, { plaintext: true });

  const res = client.invoke('authors.v1.AuthorsService/CreateAuthor', {
    name: 'TestUser',
  });

  check(res, {
    'status OK': (r) => r && r.status === grpc.StatusOK,
  });

  client.close();
  sleep(1);
};
