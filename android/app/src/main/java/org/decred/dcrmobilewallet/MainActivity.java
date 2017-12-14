package org.decred.dcrmobilewallet;

import android.content.ContextWrapper;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.view.View;
import android.widget.Button;
import android.widget.TextView;

import mobilewallet.Mobilewallet;
import mobilewallet.LibWallet;

import org.w3c.dom.Text;

public class MainActivity extends AppCompatActivity {

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        Button btn = (Button) this.findViewById((R.id.btnCreateWallet));

        ContextWrapper ctx = new ContextWrapper(this);
        final TextView txtLog = (TextView) this.findViewById(R.id.txtLog);
        final String dbDir = ctx.getFilesDir().getPath() + "/walletdb";

        final LibWallet wallet = new LibWallet(dbDir);

        btn.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                txtLog.append("Creating wallet in " + dbDir + "\n");
                try {
                    wallet.createWallet();
                    txtLog.append("Wallet Created");
                } catch (Exception e) {
                    txtLog.append("Exception creating wallet: " + e.toString());
                }
            }
        });

        Button btnOpenWallet = this.findViewById(R.id.btnOpenWallet);
        btnOpenWallet.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                txtLog.setText("Trying to open wallet\n");
                try {
                    wallet.openWallet();
                    txtLog.append("Success opening wallet\n");
                } catch (Exception e) {
                    txtLog.append("Exception opening wallet: " + e.toString());
                }
            }
        });

        final TextView txtSpendable = this.findViewById(R.id.txtSpendable);
        Button btnUpdate = this.findViewById(R.id.btnUpdateWallet);
        btnUpdate.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                try {
                    long spendable = wallet.spendableForAccount();
                    txtSpendable.setText(String.format(" %.8f", spendable / 10e7));
                } catch (Exception e) {
                    txtLog.append("Exception updating wallet: " + e.toString());
                }
            }
        });

        Button btnNewAddress = this.findViewById(R.id.btnNewAddress);
        btnNewAddress.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                try {
                    txtLog.setText("Getting address...\n");
                    String addr = wallet.addressForAccount();
                    txtLog.append("address: " + addr);
                } catch (Exception e) {
                    txtLog.append("Exception getting address: " + e.toString());
                }
            }
        });

        Button btnSend = this.findViewById(R.id.btnSend);
        btnSend.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                try {
                    txtLog.setText("Trying to send...\n");
                    String txHash = wallet.sendTx();
                    txtLog.append("Sent to faucet!\n");
                    txtLog.append(txHash);
                } catch (Exception e) {
                    txtLog.append("Exception sending: " + e.toString());
                }
            }
        });

        Button btnRescan = this.findViewById(R.id.btnRescan);
        btnRescan.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                try {
                    txtLog.setText("Rescanning...\n");
                    wallet.rescan();
                    txtLog.append("Done rescanning\n");
                } catch (Exception e) {
                    txtLog.append("Exception sending: " + e.toString());
                }
            }
        });
    }
}
