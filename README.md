# cluster-api-provider-sakuracloud

Kubernetes-native declarative infrastructure for SakuraCloud.

## What is the Cluster API Provider SakuraCloud

[Cluster API](https://github.com/kubernetes-sigs/cluster-api)のさくらのクラウド向けプロバイダー実装  

## 対応バージョン

||Cluster API v1alpha1 (v0.1)|Cluster API v1alpha2 (v0.2)|
|-|-|-|
|SakuraCloud Provider v1alpha2 (v0.1)||✓|

This provider's versions are able to install and manage the following versions of Kubernetes:

||Kubernetes 1.14|Kubernetes 1.15|Kubernetes 1.16|
|-|-|-|-|
|SakuraCloud Provider v1alpha2 (v0.1)|✓|✓|✓|

## Quick Start

- さくらのクラウド APIキーの取得/環境変数へ設定
- ソースアーカイブのビルド
- マニフェストの生成
- Management Clusterの作成
- Cluster API/Bootstrap Provider/Infrastructure Provider(さくらのクラウド向け)のデプロイ
- Control Planeの作成
- アドオン(CNIなど)のデプロイ
- Machineの作成

### さくらのクラウド APIキーの作成/環境変数へ設定

さくらのクラウド APIキーを作成し、環境変数に設定しておきます。

```bash
export SAKURACLOUD_ACCESS_TOKEN=<APIトークン>
export SAKURACLOUD_ACCESS_TOKEN_SECRET=<APIシークレット>
export SAKURACLOUD_ZONE=is1a
```

### ソースアーカイブの準備

ソースアーカイブをビルドするために`packer`,`ansible`,`qemu-img`,`usacloud`が必要です。  
事前にインストールしておいてください。

まずUbuntuのクラウドイメージをさくらのクラウドのアーカイブとしてアップロードします。

```bash
# ダウンロード(今回は18.04を利用)
curl -sL -o ubuntu.img https://cloud-images.ubuntu.com/releases/bionic/release/ubuntu-18.04-server-cloudimg-amd64.img 

# qcowからraw形式(sparse)へ変換
qemu-img convert ubuntu.img ubuntu-sparse.raw

# non-sparseファイルにする
cp --sparse=never ubuntu-sparse.raw ubuntu.raw

# さくらのクラウド上にアーカイブとしてアップロード
usacloud archive create --name cloud-image-ubuntu-18.04 --size 20 --archive-file ubuntu.raw  
```

次に、Packerでのアーカイブ作成時に利用するcloud-initのNoCloudデータソース用ISOイメージを作成してさくらのクラウドにアップロードします。

```bash
# 作業用ディレクトリ作成
mkdir workdir

# SSH公開鍵(任意のパスに置き換えてください)
export SSH_AUTHORIZED_KEY=`cat ~/.ssh/id_rsa.pub`

# user-dataファイルの作成
cat >workdir/user-data <<EOF
#cloud-config
ssh_authorized_keys: ${SSH_AUTHORIZED_KEY}
EOF

# meta-dataファイルの作成
touch workdir/meta-data

# ISOイメージの作成(linux上で作業する場合)
mkisofs -R -V cidata -o cloud-init.iso workdir/

# ISOイメージの作成)mac上で作業する場合)
# hdiutil makehybrid -iso -joliet -default-volume-name cidata -o cloud-init.iso workdir/

# さくらのクラウド上にISOファイルをアップロード
usacloud iso-image create --name packer-capi-image-builder --iso-file cloud-init.iso
```

その後Packerでアーカイブ作成を行います。

```bash
# ソースアーカイブのビルドスクリプトのリポジトリをクローン
git clone https://github.com/sacloud/image-builder.git ; cd image-builder/images/capi

# Packerプラグインのインストール
$(cd packer/sakuracloud; make plugins)

# ビルド
ARCHIVE_ID="$(usacloud archive read -q cloud-image-ubuntu-18.04)"
ISO_IMAGE_ID="$(usacloud iso-image read -q cpacker-capi-image-builder)"
PACKER_FLAGS="-var 'source_archive_id=${ARCHIVE_ID}' -var 'cloud_init_iso_image_id=${ISO_IMAGE_ID}'"
make build-sakuracloud-default 
```

デフォルトで`capi-kubernetes-template`という名前のアーカイブが作成されます。  

### マニフェストの生成

```bash
# リポジトリをクローン
git clone https://github.com/sacloud/cluster-api-provider-sakuracloud; cd cluster-api-provider-sakuracloud

# 作成するサーバに対しSSH接続したい場合は以下の環境変数を指定しておく(ユーザー名: caps)
export SSH_AUTHORIZED_KEY="$(cat ~/.ssh/id_rsa.pub)"

# マニフェストの生成(デフォルトではout/caps-exampleディレクトリ配下に出力される)
examples/generate.sh
```

### Management Clusterの作成

ここでは[kind](https://github.com/kubernetes-sigs/kind)を利用します。  
事前にインストールしておいてください。

```bash
# Management Clusterの作成
kind create cluster --name=clusterapi

# kubeconfigを作成したクラスタに向ける
export KUBECONFIG="$(kind get kubeconfig-path --name="clusterapi")"
```

### Cluster API/Bootstrap Provider/Infrastructure Provider(さくらのクラウド向け)のデプロイ

```bash
kubectl apply -f out/caps-example/provider-components.yaml
```

### Control Planeの作成

```bash
kubectl apply -f out/caps-example/cluster.yaml
kubectl apply -f out/caps-example/controlplane.yaml
```

サーバが作成されたあと、以下のコマンドで作成されたクラスタのkubeconfigを取得します。

```bash
kubectl get secret caps-example-kubeconfig -o json | jq -r .data.value | base64 --decode > caps-example.kubeconfig
```

### アドオン(CNIなど)のデプロイ

ここではCNIプラグインとしてCalicoをデプロイします。
(作成されたクラスタに対してデプロイします)

```bash
kubectl --kubeconfig=caps-example.kubeconfig apply -f out/caps-example/addons.yaml
```

### Machineの作成

```bash
kubectl -f out/caps-example/machinedeployment.yaml
```

`kubectl edit machinedeployment caps-example-md-0`コマンドで`.spec.replicas`を変更することでワーカーノードの増減が行えます。

## CRDs

### SakuraCloudCluster

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: SakuraCloudCluster
metadata:
  name: caps-example
  namespace: default
spec:
  cloudProviderConfiguration:
    zone: 'is1a'
  zone: 'is1a'
```

### SakuraCloudMachine

```yaml
# Control Plane
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: SakuraCloudMachine
metadata:
  labels:
    cluster.x-k8s.io/cluster-name: caps-example
    cluster.x-k8s.io/control-plane: "true"
  name: caps-example-controlplane-0
  namespace: default
spec:
  sourceArchive:
    filters:
    - name: Name
      values:
      - capi-kubernetes-template
  cpus: 2
  diskGB: 20
  memoryGB: 4
```

### SakuraCloudMachineTemplate

```yaml
# Worker nodes
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: SakuraCloudMachineTemplate
metadata:
  name: caps-example-md-0
  namespace: default
spec:
  template:
    spec:
      sourceArchive:
        filters:
        - name: Name
          values:
          - capi-kubernetes-template
      cpus: 2
      diskGB: 20
      memoryGB: 2
```


## License

 `cluster-api-provider-sakuracloud` Copyright (C) 2019 Kazumichi Yamamoto.

  This project is published under [Apache 2.0 License](LICENSE).
  
## Author

  * [Kazumichi Yamamoto](https://github.com/yamamoto-febc)
