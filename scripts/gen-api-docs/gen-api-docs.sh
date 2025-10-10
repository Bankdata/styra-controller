# Copyright (C) 2025 Bankdata (bankdata@bankdata.dk)

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/usr/bin/env bash

gen_styra_v1alpha1 () {
  echo "generating docs for api/styra/v1alpha1"
  ./bin/gen-crd-api-reference-docs \
    -config "scripts/gen-api-docs/config.json" \
    -api-dir "github.com/bankdata/styra-controller/api/styra/v1alpha1" \
    -template-dir "internal/template" \
    -out-file ./docs/apis/styra/v1alpha1.md
}


gen_styra_v1beta1 () {
  echo "generating docs for api/styra/v1beta1"
  ./bin/gen-crd-api-reference-docs \
    -config "scripts/gen-api-docs/config.json" \
    -api-dir "github.com/bankdata/styra-controller/api/styra/v1beta1" \
    -template-dir "internal/template" \
    -out-file ./docs/apis/styra/v1beta1.md
}

case $1 in
  styra-v1alpha1)
    gen_styra_v1alpha1
    ;;
  styra-v1beta1)
    gen_styra_v1beta1
    ;;
  all)
    gen_styra_v1alpha1
    gen_styra_v1beta1
    ;;
  *)
    echo "Usage: gen-api-docs.sh styra-v1alpha1|styra-v1beta1|all"
    ;;
esac
