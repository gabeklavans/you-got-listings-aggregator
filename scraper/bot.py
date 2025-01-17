import asyncio
import logging
import os

import telegram

logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    level=logging.INFO
)

if not os.getenv('TG_KEY') or not os.getenv('CHAT_ID'):
    logging.warn('Telegram env vars not set properly')
else:
    bot = telegram.Bot(os.getenv('TG_KEY', ''))

async def send_message(msg: str):
    if bot is None:
        return
    async with bot:
        await bot.send_message(text=msg, chat_id=int(os.getenv('CHAT_ID', '')))

def notify(url: str) -> None:
    if bot is None:
        return
    asyncio.run(send_message(f'A wild listing appeared! {url}'))
