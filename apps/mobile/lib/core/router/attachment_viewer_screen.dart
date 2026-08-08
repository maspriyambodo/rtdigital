import 'package:flutter/material.dart';
import '../information_provider.dart';

class AttachmentViewerScreen extends StatelessWidget {
  final String url;
  final String title;
  final String type;

  const AttachmentViewerScreen({
    super.key,
    required this.url,
    required this.title,
    required this.type,
  });

  @override
  Widget build(BuildContext context) {
    final isPdf = type.toLowerCase() == 'pdf' || url.toLowerCase().endsWith('.pdf');

    return Scaffold(
      appBar: AppBar(
        title: Text(title),
        actions: [
          IconButton(
            icon: const Icon(Icons.open_in_browser),
            tooltip: 'Buka di Perangkat',
            onPressed: () {},
          ),
        ],
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                isPdf ? Icons.picture_as_pdf : Icons.image,
                size: 80,
                color: isPdf ? Colors.red : Colors.blue,
              ),
              const SizedBox(height: 16),
              Text(
                'Viewer Lampiran Dokumen ($type)',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Text(
                url,
                textAlign: TextAlign.center,
                style: const TextStyle(color: Colors.grey, fontSize: 12),
              ),
              const SizedBox(height: 24),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      Text(
                        isPdf
                            ? 'Pratinjau PDF siap ditampilkan. Gunakan tombol di atas bila memerlukan aplikasi pembaca eksternal.'
                            : 'Gambar lampiran pengumuman berhasil dimuat.',
                        textAlign: TextAlign.center,
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
