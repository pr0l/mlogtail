#!/usr/bin/env python3
"""
Пример использования mlogtail HTTP API на Python
"""

import requests
import json
import time
from typing import Dict, Any


class MlogtailClient:
    """Клиент для работы с mlogtail HTTP API"""
    
    def __init__(self, base_url: str = "http://localhost:37412"):
        self.base_url = base_url.rstrip('/')
    
    def health(self) -> Dict[str, str]:
        """Проверка здоровья сервиса"""
        response = requests.get(f"{self.base_url}/health")
        response.raise_for_status()
        return response.json()
    
    def stats(self) -> Dict[str, int]:
        """Получение всех статистик"""
        response = requests.get(f"{self.base_url}/stats")
        response.raise_for_status()
        return response.json()
    
    def counter(self, name: str) -> int:
        """Получение значения конкретного счетчика"""
        response = requests.get(f"{self.base_url}/counter/{name}")
        response.raise_for_status()
        data = response.json()
        return data['value']
    
    def reset(self) -> Dict[str, str]:
        """Сброс всех счетчиков"""
        response = requests.post(f"{self.base_url}/reset")
        response.raise_for_status()
        return response.json()
    
    def stats_reset(self) -> Dict[str, int]:
        """Получение и сброс счетчиков"""
        response = requests.post(f"{self.base_url}/stats_reset")
        response.raise_for_status()
        return response.json()


def print_stats(stats: Dict[str, int]):
    """Красиво печатает статистику"""
    print("\n" + "="*50)
    print("Статистика почтового сервера Postfix")
    print("="*50)
    
    print(f"\n📨 Письма:")
    print(f"  Получено:      {stats['received']:>10,}")
    print(f"  Доставлено:    {stats['delivered']:>10,}")
    print(f"  Перенаправл.:  {stats['forwarded']:>10,}")
    print(f"  Отложено:      {stats['deferred']:>10,}")
    print(f"  Отклонено:     {stats['bounced']:>10,}")
    print(f"  Заблокировано: {stats['rejected']:>10,}")
    print(f"  Задержано:     {stats['held']:>10,}")
    print(f"  Удалено:       {stats['discarded']:>10,}")
    
    print(f"\n💾 Объем данных:")
    bytes_received_mb = stats['bytes_received'] / (1024 * 1024)
    bytes_delivered_mb = stats['bytes_delivered'] / (1024 * 1024)
    print(f"  Получено:   {bytes_received_mb:>10,.2f} MB")
    print(f"  Доставлено: {bytes_delivered_mb:>10,.2f} MB")
    
    if stats['received'] > 0:
        delivery_rate = (stats['delivered'] / stats['received']) * 100
        rejection_rate = (stats['rejected'] / stats['received']) * 100
        print(f"\n📊 Показатели:")
        print(f"  Процент доставки: {delivery_rate:>6.2f}%")
        print(f"  Процент отклонен: {rejection_rate:>6.2f}%")
    
    print("="*50 + "\n")


def monitor_continuous(client: MlogtailClient, interval: int = 5):
    """Непрерывный мониторинг с обновлением каждые N секунд"""
    print(f"Начинаю мониторинг (обновление каждые {interval} сек)...")
    print("Нажмите Ctrl+C для остановки\n")
    
    try:
        while True:
            try:
                stats = client.stats()
                print(f"\r[{time.strftime('%H:%M:%S')}] "
                      f"Получено: {stats['received']:>6} | "
                      f"Доставлено: {stats['delivered']:>6} | "
                      f"Отклонено: {stats['rejected']:>6}",
                      end='', flush=True)
                time.sleep(interval)
            except requests.RequestException as e:
                print(f"\n⚠️  Ошибка подключения: {e}")
                time.sleep(interval)
    except KeyboardInterrupt:
        print("\n\nМониторинг остановлен")


def main():
    """Основная функция"""
    client = MlogtailClient("http://localhost:37412")
    
    # Проверка здоровья
    try:
        health = client.health()
        print(f"✅ Сервис работает (версия: {health['version']})\n")
    except requests.RequestException as e:
        print(f"❌ Не удалось подключиться к сервису: {e}")
        return
    
    # Получение статистики
    stats = client.stats()
    print_stats(stats)
    
    # Примеры работы с отдельными счетчиками
    print("Примеры получения отдельных счетчиков:")
    print(f"  Получено писем: {client.counter('received')}")
    print(f"  Доставлено писем: {client.counter('delivered')}")
    print(f"  Отклонено писем: {client.counter('rejected')}")
    print()
    
    # Раскомментируйте для непрерывного мониторинга
    # monitor_continuous(client, interval=5)


if __name__ == "__main__":
    main()
