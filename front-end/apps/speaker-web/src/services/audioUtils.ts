/**
 * 音频处理工具集，负责降采样、编码与音量测算。
 */
const TARGET_SAMPLE_RATE = 16000;

export function downsampleBuffer(
  buffer: Float32Array,
  inputSampleRate: number,
  targetSampleRate = TARGET_SAMPLE_RATE
): Float32Array {
  if (targetSampleRate === inputSampleRate) {
    return buffer;
  }

  if (targetSampleRate > inputSampleRate) {
    throw new Error("目标采样率必须小于或等于输入采样率");
  }

  const ratio = inputSampleRate / targetSampleRate;
  const outputLength = Math.round(buffer.length / ratio);
  const result = new Float32Array(outputLength);
  let offsetResult = 0;
  let offsetBuffer = 0;

  while (offsetResult < result.length) {
    const nextOffsetBuffer = Math.min(Math.round((offsetResult + 1) * ratio), buffer.length);
    let accum = 0;
    let count = 0;
    for (let i = offsetBuffer; i < nextOffsetBuffer; i++) {
      accum += buffer[i];
      count++;
    }
    result[offsetResult] = count > 0 ? accum / count : 0;
    offsetResult++;
    offsetBuffer = nextOffsetBuffer;
  }

  return result;
}

export function floatTo16BitPCM(floatArray: Float32Array): Int16Array {
  const buffer = new Int16Array(floatArray.length);
  for (let i = 0; i < floatArray.length; i++) {
    const s = Math.max(-1, Math.min(1, floatArray[i]));
    buffer[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
  }
  return buffer;
}

export function pcm16ToBase64(pcm16: Int16Array): string {
  const bytes = new Uint8Array(pcm16.buffer);
  const chunkSize = 0x8000;
  let binary = "";
  for (let offset = 0; offset < bytes.length; offset += chunkSize) {
    const chunk = bytes.subarray(offset, offset + chunkSize);
    binary += String.fromCharCode(...chunk);
  }
  return btoa(binary);
}

export function calculateRMS(sample: Float32Array): number {
  let sum = 0;
  for (let i = 0; i < sample.length; i++) {
    const value = sample[i];
    sum += value * value;
  }
  const mean = sum / sample.length || 0;
  return Math.sqrt(mean);
}
