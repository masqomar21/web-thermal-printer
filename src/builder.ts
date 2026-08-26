/**
 * Chainable ESC/POS Command Builder for JavaScript & TypeScript
 * Fully mirrors Go ESCPOSBuilder hardware commands.
 */
export class ESCPOSBuilder {
  private buffer: number[] = [];

  constructor() {
    this.init();
  }

  /** Initialize printer (ESC @) */
  init(): this {
    this.buffer.push(0x1b, 0x40);
    return this;
  }

  /** Set alignment to left (ESC a 0) */
  alignLeft(): this {
    this.buffer.push(0x1b, 0x61, 0x00);
    return this;
  }

  /** Set alignment to center (ESC a 1) */
  alignCenter(): this {
    this.buffer.push(0x1b, 0x61, 0x01);
    return this;
  }

  /** Set alignment to right (ESC a 2) */
  alignRight(): this {
    this.buffer.push(0x1b, 0x61, 0x02);
    return this;
  }

  /** Set font size multiplier (1 to 8, GS ! n) */
  setFontSize(widthMulti: number = 1, heightMulti: number = 1): this {
    const w = Math.min(Math.max(widthMulti - 1, 0), 7);
    const h = Math.min(Math.max(heightMulti - 1, 0), 7);
    const n = (w << 4) | h;
    this.buffer.push(0x1d, 0x21, n);
    return this;
  }

  /** Enable or disable bold font (ESC E n) */
  setBold(enable: boolean = true): this {
    this.buffer.push(0x1b, 0x45, enable ? 0x01 : 0x00);
    return this;
  }

  /** Append text string */
  text(str: string): this {
    const encoder = new TextEncoder();
    const bytes = encoder.encode(str);
    for (let i = 0; i < bytes.length; i++) {
      this.buffer.push(bytes[i]);
    }
    return this;
  }

  /** Append text string with newline */
  textLn(str: string): this {
    return this.text(str + '\n');
  }

  /** Append empty lines */
  newLine(count: number = 1): this {
    for (let i = 0; i < count; i++) {
      this.buffer.push(0x0a);
    }
    return this;
  }

  /** Execute full paper cut with feed (GS V 65 0) */
  cutPaper(): this {
    this.buffer.push(0x1d, 0x56, 0x41, 0x00);
    return this;
  }

  /** Append horizontal divider line */
  divider(char: string = '-', lineLength: number = 32): this {
    if (!char) char = '-';
    if (lineLength <= 0) lineLength = 32;
    let line = '';
    while (line.length < lineLength) {
      line += char;
    }
    if (line.length > lineLength) {
      line = line.substring(0, lineLength);
    }
    return this.textLn(line);
  }

  /** Append two-column table row (left aligned & right aligned) */
  tableLine(left: string, right: string, totalWidth: number = 32): this {
    if (totalWidth <= 0) totalWidth = 32;
    const spaceNeeded = totalWidth - left.length - right.length;
    if (spaceNeeded <= 0) {
      return this.textLn(`${left} ${right}`);
    }
    const spaces = ' '.repeat(spaceNeeded);
    return this.textLn(`${left}${spaces}${right}`);
  }

  /** Append hardware QR Code (GS ( k) */
  qrCode(content: string, moduleSize: number = 4): this {
    if (!content) return this;
    if (moduleSize <= 0) moduleSize = 4;
    if (moduleSize > 16) moduleSize = 16;

    const encoder = new TextEncoder();
    const data = encoder.encode(content);
    const dataLen = data.length + 3;
    const pL = dataLen % 256;
    const pH = Math.floor(dataLen / 256);

    // 1. Model 2
    this.buffer.push(0x1d, 0x28, 0x6b, 0x04, 0x00, 0x31, 0x41, 0x32, 0x00);
    // 2. Size
    this.buffer.push(0x1d, 0x28, 0x6b, 0x03, 0x00, 0x31, 0x43, moduleSize);
    // 3. Error Correction L
    this.buffer.push(0x1d, 0x28, 0x6b, 0x03, 0x00, 0x31, 0x45, 0x30);
    // 4. Store Data
    this.buffer.push(0x1d, 0x28, 0x6b, pL, pH, 0x31, 0x50, 0x30);
    for (let i = 0; i < data.length; i++) {
      this.buffer.push(data[i]);
    }
    // 5. Print QR
    this.buffer.push(0x1d, 0x28, 0x6b, 0x03, 0x00, 0x31, 0x51, 0x30, 0x0a);

    return this;
  }

  /** Append CODE128 Barcode */
  barcodeCODE128(content: string): this {
    if (!content) return this;
    // GS h 60 (Height)
    this.buffer.push(0x1d, 0x68, 0x3c);
    // GS w 2 (Width)
    this.buffer.push(0x1d, 0x77, 0x02);
    // GS H 2 (Text below)
    this.buffer.push(0x1d, 0x48, 0x02);

    const encoder = new TextEncoder();
    const encoded = encoder.encode('{B' + content);
    const dataLen = encoded.length;

    this.buffer.push(0x1d, 0x6b, 0x49, dataLen);
    for (let i = 0; i < encoded.length; i++) {
      this.buffer.push(encoded[i]);
    }
    this.buffer.push(0x0a);

    return this;
  }

  /** Build and return Uint8Array byte sequence */
  toBytes(): Uint8Array {
    return new Uint8Array(this.buffer);
  }
}
