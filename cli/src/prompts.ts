import readline from "node:readline/promises";
import { stdin as input, stdout as output } from "node:process";

export async function promptText(label: string): Promise<string> {
  const rl = readline.createInterface({ input, output });
  try {
    return await rl.question(label);
  } finally {
    rl.close();
  }
}

export async function promptPassword(label: string): Promise<string> {
  if (!input.isTTY) {
    return promptText(label);
  }

  const mutableOutput = output as NodeJS.WriteStream & { muted?: boolean };
  const rl = readline.createInterface({
    input,
    output: mutableOutput,
    terminal: true,
  });
  mutableOutput.muted = true;
  const originalWrite = mutableOutput.write.bind(mutableOutput);
  mutableOutput.write = ((chunk: unknown, encoding?: BufferEncoding, callback?: (error?: Error | null) => void) => {
    if (mutableOutput.muted && typeof chunk === "string" && chunk !== label) {
      return true;
    }
    return originalWrite(chunk as string | Uint8Array, encoding as BufferEncoding, callback);
  }) as typeof mutableOutput.write;

  try {
    const answer = await rl.question(label);
    output.write("\n");
    return answer;
  } finally {
    mutableOutput.muted = false;
    mutableOutput.write = originalWrite as typeof mutableOutput.write;
    rl.close();
  }
}
