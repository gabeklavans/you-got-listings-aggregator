import os
import asyncio
import logging
from dotenv import load_dotenv
import telegram

load_dotenv()

logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    level=logging.INFO
)

bot = telegram.Bot(os.getenv('TG_KEY'))
async def send_message(msg: str):
    async with bot:
        await bot.send_message(text=msg, chat_id=int(os.getenv('CHAT_ID')))

def notify(url: str) -> None:
    asyncio.run(send_message(f'A wild listing appeared! {url}'))
