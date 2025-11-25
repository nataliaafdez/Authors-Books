import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

const client = new grpc.Client();
client.load(
  ['/Users/natalia/AuthorsBooks/proto/authorsv1'],
  'authors.proto'
);

export const options = {
  stages: [
    { duration: '30s', target: 300},  
    { duration: '30s', target: 400 },  
    { duration: '1m', target: 600},  
    { duration: '10s', target: 0 },   
  ],
  thresholds: {
    checks: ['rate>0.95'], 
  },
};

export default () => {
  const addr = __ENV.AUTHORS_ADDR;
  if (!addr) throw new Error('AUTHORS_ADDR no está definido');

  client.connect(addr, { plaintext: true });

  // 1) Crear autor
  const author = client.invoke(
    'authors.v1.AuthorsService/CreateAuthor',
    { name: `User_${Math.random()}` }
  );

  check(author, { 'create author OK': (r) => r && r.status === grpc.StatusOK });

  const authorId = author.message.author.id;

  // 2) Agregar libro (ESTRESA los 3 microservicios)
  const book = client.invoke(
    'authors.v1.AuthorsService/AddBookToAuthor',
    {
      author_id: authorId,
      title: `Libro_${Math.random()}`,
      year: 2020,
      genre: "Ficción",
      language: "ES"
    }
  );

  check(book, { 'add book OK': (r) => r && r.status === grpc.StatusOK });

  // 3) Consultar autor + libros
  const getRes = client.invoke(
    'authors.v1.AuthorsService/GetAuthor',
    { id: authorId }
  );

  check(getRes, { 'get author OK': (r) => r && r.status === grpc.StatusOK });

  client.close();
  sleep(0.5);
};
