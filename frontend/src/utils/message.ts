// Simple message utility for showing notifications
// This is a lightweight implementation that uses browser alert for now
// Can be upgraded to a toast library later if needed

type MessageType = 'success' | 'error' | 'info' | 'warning';

class MessageService {
  private showMessage(type: MessageType, content: string): void {
    // For now, use console logging and alert
    // In production, this should be replaced with a proper toast notification library
    const prefix = type.toUpperCase();
    console.log(`[${prefix}]`, content);

    // You can replace this with a toast library like react-hot-toast or sonner
    if (type === 'error') {
      alert(`Error: ${content}`);
    } else if (type === 'success') {
      // For success messages, we'll just log to console to avoid too many alerts
      // In a real app, this would show a toast notification
      console.info(`Success: ${content}`);
    }
  }

  success(content: string): void {
    this.showMessage('success', content);
  }

  error(content: string): void {
    this.showMessage('error', content);
  }

  info(content: string): void {
    this.showMessage('info', content);
  }

  warning(content: string): void {
    this.showMessage('warning', content);
  }
}

export const message = new MessageService();
