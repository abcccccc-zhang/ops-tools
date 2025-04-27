import axios from 'axios';

export const decryptMessage = async (encrypted_string, encryption_key,algorithm) => {
  const response = await axios.post('/api/decrypt', {
    encrypted_string,
    encryption_key,
    algorithm,
  });
  return response.data;
};
