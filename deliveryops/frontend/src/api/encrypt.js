import axios from 'axios';

export const encryptMessage = async (msg, encryption_key,algorithm) => {
  const response = await axios.post('/api/encrypt', {
    msg,
    encryption_key,
    algorithm,
  });
  return response.data;
};
